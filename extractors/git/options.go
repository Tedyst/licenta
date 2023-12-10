package git

import (
	"context"
	"log/slog"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/tedyst/licenta/extractors/file"
)

const defaultProbability = 0.7

type Option func(*options) error

type options struct {
	credentials        transport.AuthMethod
	probability        float64
	ignoreFileNames    []string
	fileScannerOptions []file.Option
	skipCommitFunc     func(batch []BatchItem) ([]BatchItem, error)
	callbackResult     func(ctx context.Context, scanner *GitScan, result *GitResult) error
}

func WithCredentials(creds transport.AuthMethod) Option {
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

func WithSkipCommitFunc(f func(batch []BatchItem) ([]BatchItem, error)) Option {
	return func(o *options) error {
		o.skipCommitFunc = f
		return nil
	}
}

func WithCallbackResult(f func(ctx context.Context, scanner *GitScan, result *GitResult) error) Option {
	return func(o *options) error {
		o.callbackResult = f
		return nil
	}
}

func makeOptions(opts ...Option) (*options, error) {
	options := &options{
		probability:     defaultProbability,
		ignoreFileNames: defaultIgnoreFileNameIncluding[:],
		callbackResult: func(ctx context.Context, scanner *GitScan, result *GitResult) error {
			for _, r := range result.Results {
				slog.InfoContext(ctx, "Found Git result", "filename", result.FileName, "commit_hash", result.CommitHash, "line", r.Line, "username", r.Username, "password", r.Password, "probability", r.Probability)
			}
			return nil
		},
		skipCommitFunc: func(batch []BatchItem) ([]BatchItem, error) {
			return batch, nil
		},
	}
	for _, opt := range opts {
		err := opt(options)
		if err != nil {
			return nil, err
		}
	}
	return options, nil
}
