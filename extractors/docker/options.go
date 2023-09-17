package docker

import (
	"runtime"

	"github.com/google/go-containerregistry/pkg/authn"
)

type Option func(*options) error

type options struct {
	credentials     authn.Authenticator
	concurrency     int
	probability     float32
	ignoreFileNames []string
}

func WithCredentials(creds authn.Authenticator) Option {
	return func(o *options) error {
		o.credentials = creds
		return nil
	}
}

func WithConcurrency(concurrency int) Option {
	return func(o *options) error {
		o.concurrency = concurrency
		return nil
	}
}

func WithProbability(probability float32) Option {
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

func makeOptions(opts ...Option) (*options, error) {
	o := &options{
		concurrency: runtime.NumCPU(),
		probability: 0.7,
	}
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}
