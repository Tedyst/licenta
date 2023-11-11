package docker

import (
	"archive/tar"
	"context"
	"io"
	"strings"
	"sync"

	"github.com/djherbis/buffer"
	"github.com/djherbis/nio"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/extractors/file"
	"golang.org/x/exp/slog"
)

type LayerResult struct {
	Layer    string
	FileName string
	Results  []file.ExtractResult
}

type DockerScan struct {
	options       *options
	reference     name.Reference
	scannedLayers []v1.Layer
	errorChannel  chan error
}

func (scanner *DockerScan) scanFile(ctx context.Context, reader io.Reader, header tar.Header, layer string) error {
	results, err := file.ExtractFromReader(ctx, header.Name, reader, scanner.options.fileScannerOptions...)
	if err != nil {
		return err
	}

	if len(results) > 0 {
		err := scanner.options.callbackResult(scanner, &LayerResult{
			Layer:    layer,
			FileName: header.Name,
			Results:  results,
		})
		if err != nil {
			return errors.Wrap(err, "worker: cannot callback result")
		}
	}
	return nil
}

func isFileNameIgnored(name string) bool {
	for _, ignore := range defaultIgnoreFileNameIncluding {
		if strings.Contains(name, ignore) {
			return true
		}
	}
	return false
}

func (scanner *DockerScan) scanTarArchive(ctx context.Context, archive tar.Reader, digest string) error {
	wg := &sync.WaitGroup{}

	for {
		header, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrap(err, "scanTarArchive: failed to read file from archive")
		}

		if !header.FileInfo().Mode().IsRegular() {
			continue
		}

		if isFileNameIgnored(header.Name) {
			continue
		}

		b := buffer.New(32 * 1024)
		r, w := nio.Pipe(b)

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer r.Close()
			err := scanner.scanFile(ctx, r, *header, digest)
			if err != nil {
				scanner.errorChannel <- err
			}
		}()

		_, err = io.Copy(w, &archive)
		w.Close()
		if err != nil && !errors.Is(err, io.ErrClosedPipe) {
			return errors.Wrap(err, "scanTarArchive: failed to read file from archive using Copy")
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	waitChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case err := <-scanner.errorChannel:
		return errors.Wrap(err, "scanTarArchive: caught error worker")
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "scanTarArchive: context is done")
	case <-waitChan:
		slog.InfoContext(ctx, "scanTarArchive: finished processing archive", "digest", digest)
	}
	return nil
}

func (scanner *DockerScan) processLayer(ctx context.Context, layer v1.Layer) error {
	slog.DebugContext(ctx, "processLayer: processing layer", "layer", layer)

	digest, err := layer.Digest()
	if err != nil {
		return errors.Wrap(err, "processLayer: cannot get digest for layer")
	}
	reader, err := layer.Uncompressed()
	if err != nil {
		return errors.Wrap(err, "processLayer: cannot get layer reader")
	}
	defer reader.Close()

	tarReader := tar.NewReader(reader)
	return scanner.scanTarArchive(ctx, *tarReader, digest.String())
}

func NewScanner(ctx context.Context, imageName string, opts ...Option) (*DockerScan, error) {
	slog.InfoContext(ctx, "NewScanner: creating new scanner", "image", imageName)

	scanner := &DockerScan{}

	o, err := makeOptions(opts...)
	if err != nil {
		return nil, errors.Wrap(err, "NewScanner: cannot make options")
	}

	scanner.options = o

	ref, err := name.ParseReference(imageName)
	if err != nil {
		return nil, errors.Wrap(err, "NewScanner: cannot parse reference")
	}

	scanner.reference = ref

	scanner.errorChannel = make(chan error)

	slog.InfoContext(ctx, "NewScanner: finished creating new scanner", "image", imageName)
	return scanner, nil
}

func (scanner *DockerScan) FindLayers(ctx context.Context) ([]v1.Layer, error) {
	result := []v1.Layer{}
	index, err := remote.Index(scanner.reference, remote.WithAuth(scanner.options.credentials), remote.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "ProcessImage: cannot get index for image")
	}
	indexManifest, err := index.IndexManifest()
	if err != nil {
		return nil, errors.Wrap(err, "ProcessImage: cannot get index manifest for image")
	}

	for _, manifest := range indexManifest.Manifests {
		img, err := index.Image(manifest.Digest)
		if err != nil {
			return nil, errors.Wrap(err, "ProcessImage: cannot get image from digest")
		}
		layers, err := img.Layers()
		if err != nil {
			return nil, errors.Wrap(err, "ProcessImage: cannot get layers for image")
		}
		result = append(result, layers...)
	}

	return result, nil
}

func (scanner *DockerScan) ScannedLayers() []string {
	result := []string{}
	for _, l := range scanner.scannedLayers {
		digest, err := l.Digest()
		if err != nil {
			continue
		}
		result = append(result, digest.String())
	}
	return result
}

func (scanner *DockerScan) ProcessLayers(ctx context.Context, layers []v1.Layer) error {
	slog.InfoContext(ctx, "ProcessLayers: processing layers")

	ctx, cancelCtx := context.WithTimeout(ctx, scanner.options.timeout)
	defer cancelCtx()

	waitCh := make(chan struct{})

	wg := sync.WaitGroup{}
	wg.Add(len(layers))
	for _, layer := range layers {
		layer := layer
		go func() {
			defer wg.Done()
			err := scanner.processLayer(ctx, layer)
			if err != nil {
				scanner.errorChannel <- err
			}
		}()
	}

	go func() {
		wg.Wait()
		close(waitCh)
	}()

	select {
	case err := <-scanner.errorChannel:
		return errors.Wrap(err, "ProcessLayers: caught error worker")
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "ProcessLayers: context is done")
	case <-waitCh:
		slog.InfoContext(ctx, "ProcessLayers: finished processing layers")
		return nil
	}
}
