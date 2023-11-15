package git

import (
	"bufio"
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/extractors/file"
)

func (scanner *GitScan) inspectBinaryFile(ctx context.Context, commit *object.Commit, to diff.File) ([]file.ExtractResult, error) {
	content, err := commit.File(to.Path())
	if err == object.ErrFileNotFound {
		// File was renamed, this is not supported yet by go-git
		return nil, nil
	}
	if err != nil {
		println(commit.Hash.String())
		return nil, errors.Wrap(err, "inspectBinaryFile: cannot get file contents")
	}
	rd, err := content.Blob.Reader()
	if err != nil {
		return nil, errors.Wrap(err, "inspectBinaryFile: cannot open reader")
	}
	defer rd.Close()

	results, err := file.ExtractFromReader(ctx, to.Path(), rd, scanner.options.fileScannerOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "inspectBinaryFile: cannot extract from reader")
	}
	return results, nil
}

func (scanner *GitScan) inspectTextFile(ctx context.Context, filePatch diff.FilePatch) ([]file.ExtractResult, error) {
	var results []file.ExtractResult
	var lineNumber int = 0
	for _, chunk := range filePatch.Chunks() {
		switch chunk.Type() {
		case diff.Equal:
			lineNumber += strings.Count(chunk.Content(), "\n")
		case diff.Add:
			var scanner = bufio.NewScanner(strings.NewReader(chunk.Content()))
			for scanner.Scan() {
				lineNumber++
				line := scanner.Text()
				_, toFile := filePatch.Files()
				fileResults, err := file.ExtractFromLine(ctx, toFile.Path(), lineNumber, line)
				if err != nil {
					return nil, errors.Wrap(err, "inspectTextFile: cannot extract from line")
				}
				results = append(results, fileResults...)
			}
		}
	}
	return results, nil
}

func (scanner *GitScan) inspectFilePatch(ctx context.Context, commit *object.Commit, filePatch diff.FilePatch) error {
	var results []file.ExtractResult
	var err error
	_, to := filePatch.Files()
	if to == nil {
		return nil
	}
	if filePatch.IsBinary() {
		results, err = scanner.inspectBinaryFile(ctx, commit, to)
	} else {
		results, err = scanner.inspectTextFile(ctx, filePatch)
	}
	if err != nil {
		return err
	}

	results = file.FilterExtractResultsByProbability(ctx, results, scanner.options.probability)

	if len(results) > 0 {
		err = scanner.options.callbackResult(scanner, &GitResult{
			CommitHash: commit.Hash.String(),
			FileName:   to.Path(),
			Results:    results,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (scanner *GitScan) inspectCommit(ctx context.Context, commit *object.Commit, parent *object.Commit) error {
	scanner.mutex.Lock()
	patch, err := commit.Patch(parent)
	if err != nil {
		scanner.mutex.Unlock()
		return err
	}
	scanner.mutex.Unlock()

	parentHash := ""
	if parent != nil {
		parentHash = parent.Hash.String()
	}
	slog.InfoContext(ctx, "Inspecting commit", "commit", commit.Hash.String(), "parent", parentHash)

	wg := sync.WaitGroup{}
	errorChannel := make(chan error)
	waitChannel := make(chan struct{})

	for _, filePatch := range patch.FilePatches() {
		filePatch := filePatch
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := scanner.inspectFilePatch(ctx, commit, filePatch)
			if err != nil {
				errorChannel <- err
			}
		}()
	}

	go func() {
		wg.Wait()
		close(waitChannel)
	}()

	select {
	case err := <-errorChannel:
		return errors.Wrap(err, "inspectCommit: caught error from worker")
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "inspectCommit: context is done")
	case <-waitChannel:
		return nil
	}
}

type BatchItem struct {
	Commit *object.Commit
	Parent *object.Commit
}

func (scanner *GitScan) Scan(ctx context.Context) error {
	objIter, err := scanner.repository.CommitObjects()
	if err != nil {
		return err
	}

	errorChannel := make(chan error)
	waitChannel := make(chan struct{})
	wg := sync.WaitGroup{}

	batch := []BatchItem{}

	err = objIter.ForEach(func(c *object.Commit) error {
		parentIter := c.Parents()
		hasParent := false
		err := parentIter.ForEach(func(parent *object.Commit) error {
			hasParent = true
			batch = append(batch, BatchItem{
				Commit: c,
				Parent: parent,
			})
			return nil
		})
		if err != nil {
			return err
		}
		if !hasParent {
			batch = append(batch, BatchItem{
				Commit: c,
				Parent: nil,
			})
		}

		if len(batch) >= 50 {
			batch, err = scanner.options.skipCommitFunc(batch)
			if err != nil {
				return err
			}
			for _, obj := range batch {
				o := obj
				wg.Add(1)
				go func() {
					defer wg.Done()
					if ctx.Err() != nil {
						return
					}
					err := scanner.inspectCommit(ctx, o.Commit, o.Parent)
					if err != nil {
						errorChannel <- err
					}
				}()
			}
			batch = []BatchItem{}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "sad")
	}

	for _, obj := range batch {
		obj := obj
		wg.Add(1)
		go func() {
			defer wg.Done()
			if ctx.Err() != nil {
				return
			}
			err := scanner.inspectCommit(ctx, obj.Commit, obj.Parent)
			if err != nil {
				errorChannel <- err
			}
		}()
	}

	go func() {
		wg.Wait()
		close(waitChannel)
	}()

	select {
	case err := <-errorChannel:
		return errors.Wrap(err, "Scan: caught error from worker")
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "Scan: context is done")
	case <-waitChannel:
		return nil
	}
}
