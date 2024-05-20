package git

import (
	"bufio"
	"context"
	errorss "errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"errors"

	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/tedyst/licenta/extractors/file"
)

const maxPreviousLines = 5

func (scanner *GitScan) inspectBinaryFile(ctx context.Context, commit *object.Commit, to diff.File) (results []file.ExtractResult, err error) {
	scanner.mutex.Lock()
	content, err := commit.File(to.Path())
	if err == object.ErrFileNotFound {
		scanner.mutex.Unlock()
		// File was renamed, this is not supported yet by go-git
		return nil, nil
	}
	if err != nil {
		scanner.mutex.Unlock()
		return nil, fmt.Errorf("inspectBinaryFile: cannot get file contents: %w", err)
	}
	rd, err := content.Blob.Reader()
	if err != nil {
		scanner.mutex.Unlock()
		return nil, fmt.Errorf("inspectBinaryFile: cannot open reader: %w", err)
	}
	defer func() {
		err = errorss.Join(err, rd.Close())
	}()
	scanner.mutex.Unlock()

	results, err = scanner.fileScanner.ExtractFromReader(ctx, to.Path(), rd)
	if err != nil {
		return nil, fmt.Errorf("inspectBinaryFile: cannot extract from reader: %w", err)
	}
	return results, nil
}

func (scanner *GitScan) inspectTextFile(ctx context.Context, filePatch diff.FilePatch) ([]file.ExtractResult, error) {
	var results []file.ExtractResult
	var lineNumber int = 0
	var previousLines []string
	for _, chunk := range filePatch.Chunks() {
		switch chunk.Type() {
		case diff.Equal:
			lineNumber += strings.Count(chunk.Content(), "\n")
		case diff.Add:
			var sc = bufio.NewScanner(strings.NewReader(chunk.Content()))
			for sc.Scan() {
				lineNumber++
				line := sc.Text()
				_, toFile := filePatch.Files()
				fileResults, err := scanner.fileScanner.ExtractFromLine(ctx, toFile.Path(), lineNumber, line, strings.Join(previousLines, "\n"))
				if err != nil {
					return nil, fmt.Errorf("inspectTextFile: cannot extract from line: %w", err)
				}
				results = append(results, fileResults...)
			}
		}

		previousLines = append(previousLines, strings.Split(chunk.Content(), "\n")...)
		if len(previousLines) > maxPreviousLines {
			previousLines = previousLines[len(previousLines)-maxPreviousLines:]
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
		return fmt.Errorf("inspectFilePatch: cannot inspect file: %w", err)
	}

	scanner.mutex.Lock()
	defer scanner.mutex.Unlock()

	foundResults.Add(ctx, int64(len(results)))
	err = scanner.options.callbackResult(ctx, scanner, &GitResult{
		Commit:   commit,
		FileName: to.Path(),
		Results:  results,
	})
	if err != nil {
		return fmt.Errorf("inspectFilePatch: cannot callback result: %w", err)
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

	slog.DebugContext(ctx, "Inspecting commit", "commit", commit.Hash.String(), "parent", parentHash)
	commitsInspected.Add(ctx, 1)

	wg := sync.WaitGroup{}
	errorChannel := make(chan error)
	waitChannel := make(chan struct{})

	err = scanner.options.callbackResult(ctx, scanner, &GitResult{
		Commit:   commit,
		FileName: "",
		Results:  []file.ExtractResult{},
	})
	if err != nil {
		return fmt.Errorf("inspectFilePatch: cannot create commit: %w", err)
	}

	for _, filePatch := range patch.FilePatches() {
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
		return fmt.Errorf("inspectCommit: caught error from worker: %w", err)
	case <-ctx.Done():
		return fmt.Errorf("inspectCommit: context is done: %w", ctx.Err())
	case <-waitChannel:
		return nil
	}
}

type BatchItem struct {
	Commit *object.Commit
	Parent *object.Commit
}

func (scanner *GitScan) Scan(ctx context.Context) error {
	if !scanner.initiated {
		return errors.New("not initiated")
	}

	ctx, span := tracer.Start(ctx, "GitScan.Scan")
	defer span.End()

	objIter, err := scanner.repository.CommitObjects()
	if err != nil {
		return err
	}

	errorChannel := make(chan error)
	waitChannel := make(chan struct{})
	wg := sync.WaitGroup{}

	batch := []BatchItem{}

	scanner.mutex.Lock()
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
			return fmt.Errorf("Scan: cannot iterate over parents: %w", err)
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
		scanner.mutex.Unlock()
		return fmt.Errorf("Scan: cannot iterate over commits: %w", err)
	}

	batch, err = scanner.options.skipCommitFunc(batch)
	if err != nil {
		scanner.mutex.Unlock()
		return err
	}
	for _, obj := range batch {
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

	scanner.mutex.Unlock()

	go func() {
		wg.Wait()
		close(waitChannel)
	}()

	select {
	case err := <-errorChannel:
		return fmt.Errorf("Scan: caught error from worker: %w", err)
	case <-ctx.Done():
		return fmt.Errorf("Scan: context is done: %w", ctx.Err())
	case <-waitChannel:
		return nil
	}
}
