package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"strings"
	"sync"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/extractors/file"
	"golang.org/x/exp/slog"
)

type channelTask struct {
	layer    string
	fileName string
	content  []byte
}

type LayerResult struct {
	Layer    string
	FileName string
	Results  []file.ExtractResult
}

func scanFileWorker(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup, channel chan *channelTask, callbackResult func(result LayerResult) error, o *options) error {
	for {
		task, ok := <-channel
		if !ok || task == nil || task.content == nil {
			wg.Done()
			return nil
		}

		slog.DebugContext(ctx, "scanFile: scanning file", "layer", task.layer, "file", task.fileName)

		results, err := file.ExtractFromReader(ctx, task.fileName, bytes.NewReader(task.content), o.fileScannerOptions...)
		if err != nil {
			cancelFunc()
			return errors.Wrap(err, "scanFileWorker: cannot extract from reader")
		}
		results = file.FilterExtractResultsByProbability(ctx, results, o.probability)

		slog.DebugContext(ctx, "scanFile: finished scanning file", "layer", task.layer, "file", task.fileName)

		if len(results) > 0 {
			err := callbackResult(LayerResult{
				Layer:    task.layer,
				FileName: task.fileName,
				Results:  results,
			})
			if err != nil {
				cancelFunc()
				return errors.Wrap(err, "scanFileWorker: cannot callback result")
			}
		}

		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "scanFile")
		default:
		}
	}
}

func isFileNameIgnored(name string) bool {
	for _, ignore := range defaultIgnoreFileNameIncluding {
		if strings.Contains(name, ignore) {
			return true
		}
	}
	return false
}

func processLayer(ctx context.Context, c chan *channelTask, layer v1.Layer) error {
	slog.DebugContext(ctx, "processLayer: processing layer", "layer", layer)

	digest, err := layer.Digest()
	if err != nil {
		return errors.Wrap(err, "processLayer: cannot get digest for layer")
	}
	reader, err := layer.Uncompressed()
	if err != nil {
		return errors.Wrap(err, "processLayer: cannot get layer reader")
	}
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrap(err, "processLayer: failed to read file from layer")
		}

		if !header.FileInfo().Mode().IsRegular() {
			continue
		}

		if isFileNameIgnored(header.Name) {
			continue
		}

		content := make([]byte, header.Size)
		b, err := tarReader.Read(content)
		if b == 0 && err == io.EOF {
			continue
		}
		if err != nil && err != io.EOF {
			return errors.Wrap(err, "processLayer: failed to read layer file content")
		}

		c <- &channelTask{
			layer:    digest.String(),
			fileName: header.Name,
			content:  content,
		}
	}

	slog.DebugContext(ctx, "processLayer: finished processing layer", "layer", layer)
	return nil
}

func processImage(ctx context.Context, c chan *channelTask, image v1.Image, opt *options) error {
	slog.DebugContext(ctx, "processImage: processing image", "image", image)

	layers, err := image.Layers()
	if err != nil {
		return errors.Wrap(err, "processImage: cannot get layers for image")
	}
	for _, layer := range layers {
		digest, err := layer.Digest()
		if err != nil {
			return errors.Wrap(err, "processImage: cannot get digest")
		}
		if opt.skipLayer != nil && opt.skipLayer(digest.String()) {
			slog.DebugContext(ctx, "processImage: skipping layer", "layer", digest)
			continue
		}
		err = processLayer(ctx, c, layer)
		if err != nil {
			return errors.Wrap(err, "processImage: cannot process layer")
		}

		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "processImage: context deadline exceeded")
		default:
		}
	}

	slog.DebugContext(ctx, "processImage: finished processing image", "image", image)
	return nil
}

func ProcessImage(
	ctx context.Context,
	imageName string,
	callbackResult func(result LayerResult) error,
	opts ...Option,
) error {
	slog.InfoContext(ctx, "ProcessImage: processing image", "image", imageName, "opts", opts)

	o, err := makeOptions(opts...)
	if err != nil {
		return errors.Wrap(err, "ProcessImage: cannot make options")
	}

	ctx, cancelCtx := context.WithTimeout(ctx, o.timeout)

	ref, err := name.ParseReference(imageName)
	if err != nil {
		cancelCtx()
		return errors.Wrap(err, "ProcessImage: cannot parse reference")
	}

	index, err := remote.Index(ref, remote.WithAuth(o.credentials), remote.WithContext(ctx))
	if err != nil {
		cancelCtx()
		return errors.Wrap(err, "ProcessImage: cannot get index for image")
	}
	indexManifest, err := index.IndexManifest()
	if err != nil {
		cancelCtx()
		return errors.Wrap(err, "ProcessImage: cannot get index manifest for image")
	}

	resultChan := make(chan *channelTask, o.concurrency)
	var wg sync.WaitGroup
	wg.Add(o.concurrency)

	for i := 0; i < o.concurrency; i++ {
		go scanFileWorker(ctx, cancelCtx, &wg, resultChan, callbackResult, o)
	}

	for _, manifest := range indexManifest.Manifests {
		img, err := index.Image(manifest.Digest)
		if err != nil {
			cancelCtx()
			return errors.Wrap(err, "ProcessImage: cannot get image from digest")
		}
		processImage(ctx, resultChan, img, o)
	}

	close(resultChan)
	wg.Wait()
	cancelCtx()

	slog.InfoContext(ctx, "ProcessImage: finished processing image", "image", imageName, "opts", opts)
	return nil
}
