package docker

import (
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/tedyst/licenta/extractors/file"
	"golang.org/x/exp/slog"
)

type Option func(*options) error

type options struct {
	credentials        authn.Authenticator
	probability        float64
	ignoreFileNames    []string
	fileScannerOptions []file.Option
	timeout            time.Duration
	skipLayerFunc      func(layer string) bool
	callbackResult     func(scanner *DockerScan, result *LayerResult) error
}

func WithCallbackResult(f func(scanner *DockerScan, result *LayerResult) error) Option {
	return func(o *options) error {
		o.callbackResult = f
		return nil
	}
}

func WithSkipLayer(f func(layer string) bool) Option {
	return func(o *options) error {
		o.skipLayerFunc = f
		return nil
	}
}

func WithCredentials(creds authn.Authenticator) Option {
	return func(o *options) error {
		o.credentials = creds
		return nil
	}
}

func WithProbability(probability float64) Option {
	return func(o *options) error {
		o.probability = probability
		return nil
	}
}

func WithIgnoreFileNames(useDefault bool, names ...string) Option {
	return func(o *options) error {
		if useDefault {
			o.ignoreFileNames = defaultIgnoreFileNameIncluding[:]
		}
		o.ignoreFileNames = append(o.ignoreFileNames, names...)
		return nil
	}
}

func WithFileScannerOptions(opts ...file.Option) Option {
	return func(o *options) error {
		o.fileScannerOptions = append(o.fileScannerOptions, opts...)
		return nil
	}
}

func makeOptions(opts ...Option) (*options, error) {
	o := &options{
		probability: 0.7,
		timeout:     time.Hour,
		callbackResult: func(scanner *DockerScan, result *LayerResult) error {
			slog.Info("ProcessLayers: layer result", "layer", result.Layer, "result", result)
			return nil
		},
	}
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}
