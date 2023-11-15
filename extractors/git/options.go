package git

import (
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
	branch             string
	skipCommitFunc     func(batch []BatchItem) ([]BatchItem, error)
	callbackResult     func(scanner *GitScan, result *GitResult) error
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

func WithBranch(branch string) Option {
	return func(o *options) error {
		o.branch = branch
		return nil
	}
}

func WithSkipCommitFunc(f func(batch []BatchItem) ([]BatchItem, error)) Option {
	return func(o *options) error {
		o.skipCommitFunc = f
		return nil
	}
}

func makeOptions(opts ...Option) (*options, error) {
	options := &options{
		probability:     defaultProbability,
		ignoreFileNames: defaultIgnoreFileNameIncluding[:],
		branch:          "master",
		callbackResult: func(scanner *GitScan, result *GitResult) error {
			slog.Info("ProcessGit: git result", "result", result, "scanner", scanner)
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
