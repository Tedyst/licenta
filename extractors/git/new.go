package git

import (
	"sync"

	gitgo "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/tedyst/licenta/extractors/file"
)

type GitResult struct {
	CommitHash string
	FileName   string
	Results    []file.ExtractResult
}

type GitScan struct {
	options    *options
	repository *gitgo.Repository

	mutex sync.Mutex
}

func NewFromRepo(repository *gitgo.Repository, options ...Option) (*GitScan, error) {
	o, err := makeOptions(options...)
	if err != nil {
		return nil, err
	}
	return &GitScan{
		options:    o,
		repository: repository,
	}, nil
}

func New(repoUrl string, options ...Option) (*GitScan, error) {
	o, err := makeOptions(options...)
	if err != nil {
		return nil, err
	}
	repository, err := gitgo.Clone(memory.NewStorage(), nil, &gitgo.CloneOptions{
		URL:  repoUrl,
		Auth: o.credentials,
	})
	if err != nil {
		return nil, err
	}
	return &GitScan{
		options:    o,
		repository: repository,
	}, nil
}
