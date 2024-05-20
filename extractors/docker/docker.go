package docker

import (
	"archive/tar"
	"context"
	errorss "errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"log/slog"

	"errors"

	"github.com/djherbis/buffer"
	"github.com/djherbis/nio/v3"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/tedyst/licenta/extractors/file"
)

type FileScanner interface {
	ExtractFromReader(ctx context.Context, fileName string, reader io.Reader) ([]file.ExtractResult, error)
}

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
	fileScanner   FileScanner
}

func (scanner *DockerScan) scanFile(ctx context.Context, reader io.Reader, header tar.Header, layer string) error {
	results, err := scanner.fileScanner.ExtractFromReader(ctx, header.Name, reader)
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
			return fmt.Errorf("worker: cannot callback result: %w", err)
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

func (scanner *DockerScan) scanTarArchive(ctx context.Context, archive tar.Reader, layer v1.Layer) error {
	wg := &sync.WaitGroup{}

	digest, err := layer.Digest()
	if err != nil {
		return fmt.Errorf("processLayer: cannot get digest for layer: %w", err)
	}

	for {
		header, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("scanTarArchive: failed to read file from archive: %w", err)
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
			defer func() {
				err := r.Close()
				if err != nil {
					scanner.errorChannel <- err
				}
			}()
			err := scanner.scanFile(ctx, r, *header, digest.String())
			if err != nil {
				scanner.errorChannel <- err
			}
		}()

		_, err = io.Copy(w, &archive)
		if err != nil && !errors.Is(err, io.ErrClosedPipe) {
			err2 := w.Close()
			return errorss.Join(fmt.Errorf("scanTarArchive: failed to read file from archive using Copy: %w", err), err2)
		}
		err = w.Close()
		if err != nil {
			return fmt.Errorf("scanTarArchive: failed to close writer: %w", err)
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
		return fmt.Errorf("scanTarArchive: caught error worker: %w", err)
	case <-ctx.Done():
		return fmt.Errorf("scanTarArchive: context is done: %w", ctx.Err())
	case <-waitChan:
		scanner.scannedLayers = append(scanner.scannedLayers, layer)
		slog.InfoContext(ctx, "scanTarArchive: finished processing archive", "digest", digest)
	}
	return nil
}

func (scanner *DockerScan) processLayer(ctx context.Context, layer v1.Layer) (err error) {
	slog.DebugContext(ctx, "processLayer: processing layer", "layer", layer)

	reader, err := layer.Uncompressed()
	if err != nil {
		return fmt.Errorf("processLayer: cannot get layer reader: %w", err)
	}
	defer func() {
		err = errorss.Join(err, reader.Close())
	}()

	tarReader := tar.NewReader(reader)
	return scanner.scanTarArchive(ctx, *tarReader, layer)
}

func NewScanner(ctx context.Context, fileScanner FileScanner, imageName string, opts ...Option) (*DockerScan, error) {
	slog.InfoContext(ctx, "NewScanner: creating new scanner", "image", imageName)

	scanner := &DockerScan{
		fileScanner: fileScanner,
	}

	o, err := makeOptions(opts...)
	if err != nil {
		return nil, fmt.Errorf("NewScanner: cannot make options: %w", err)
	}

	scanner.options = o

	ref, err := name.ParseReference(imageName)
	if err != nil {
		return nil, fmt.Errorf("NewScanner: cannot parse reference: %w", err)
	}

	scanner.reference = ref

	scanner.errorChannel = make(chan error)

	return scanner, nil
}

func (scanner *DockerScan) FindLayers(ctx context.Context) ([]v1.Layer, error) {
	result := []v1.Layer{}
	index, err := remote.Index(scanner.reference, remote.WithAuth(scanner.options.credentials), remote.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("ProcessImage: cannot get index for image: %w", err)
	}
	indexManifest, err := index.IndexManifest()
	if err != nil {
		return nil, fmt.Errorf("ProcessImage: cannot get index manifest for image: %w", err)
	}

	for _, manifest := range indexManifest.Manifests {
		img, err := index.Image(manifest.Digest)
		if err != nil {
			return nil, fmt.Errorf("ProcessImage: cannot get image from digest: %w", err)
		}
		layers, err := img.Layers()
		if err != nil {
			return nil, fmt.Errorf("ProcessImage: cannot get layers for image: %w", err)
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
		return fmt.Errorf("ProcessLayers: caught error worker: %w", err)
	case <-ctx.Done():
		return fmt.Errorf("ProcessLayers: context is done: %w", ctx.Err())
	case <-waitCh:
		slog.InfoContext(ctx, "ProcessLayers: finished processing layers")
		return nil
	}
}
