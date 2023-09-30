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

func scanFileWorker(ctx context.Context, wg *sync.WaitGroup, channel chan *channelTask, callbackResult func(result LayerResult), o *options) error {
	for {
		task, ok := <-channel
		if !ok || task == nil || task.content == nil {
			wg.Done()
			return nil
		}

		slog.DebugContext(ctx, "scanFile: scanning file", "layer", task.layer, "file", task.fileName)

		results, err := file.ExtractFromReader(ctx, task.fileName, bytes.NewReader(task.content), o.fileScannerOptions...)
		if err != nil {
			return err
		}
		results = file.FilterExtractResultsByProbability(ctx, results, o.probability)

		slog.DebugContext(ctx, "scanFile: finished scanning file", "layer", task.layer, "file", task.fileName)

		if len(results) > 0 {
			callbackResult(LayerResult{
				Layer:    task.layer,
				FileName: task.fileName,
				Results:  results,
			})
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
		return err
	}
	reader, err := layer.Uncompressed()
	if err != nil {
		return err
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

func processImage(ctx context.Context, c chan *channelTask, image v1.Image) error {
	slog.DebugContext(ctx, "processImage: processing image", "image", image)

	layers, err := image.Layers()
	if err != nil {
		return err
	}
	for _, layer := range layers {
		err = processLayer(ctx, c, layer)
		if err != nil {
			return err
		}
	}

	slog.DebugContext(ctx, "processImage: finished processing image", "image", image)
	return nil
}

func ProcessImage(
	ctx context.Context,
	imageName string,
	callbackResult func(result LayerResult),
	opts ...Option,
) error {
	slog.InfoContext(ctx, "ProcessImage: processing image", "image", imageName, "opts", opts)

	o, err := makeOptions(opts...)
	if err != nil {
		return err
	}

	ref, err := name.ParseReference(imageName)
	if err != nil {
		return err
	}

	index, err := remote.Index(ref, remote.WithAuth(o.credentials), remote.WithContext(ctx))
	if err != nil {
		return err
	}
	indexManifest, err := index.IndexManifest()
	if err != nil {
		return err
	}

	resultChan := make(chan *channelTask, o.concurrency)
	var wg sync.WaitGroup
	wg.Add(o.concurrency)

	for i := 0; i < o.concurrency; i++ {
		go scanFileWorker(ctx, &wg, resultChan, callbackResult, o)
	}

	for _, manifest := range indexManifest.Manifests {
		img, err := index.Image(manifest.Digest)
		if err != nil {
			return err
		}
		processImage(ctx, resultChan, img)
	}

	close(resultChan)
	wg.Wait()

	slog.InfoContext(ctx, "ProcessImage: finished processing image", "image", imageName, "opts", opts)
	return nil
}
