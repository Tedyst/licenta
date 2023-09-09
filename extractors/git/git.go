package git

import (
	"bufio"
	"fmt"
	"strings"

	gitgo "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/tedyst/licenta/extractors/file"
)

func inspectCommit(commit *object.Commit, parent *object.Commit, cutoffProbability float32) ([]file.ExtractResult, error) {
	patch, err := commit.Patch(parent)
	if err != nil {
		return nil, err
	}
	var results []file.ExtractResult
	for _, filePatch := range patch.FilePatches() {
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
					results = append(results, file.ExtractFromLine(toFile.Path(), lineNumber, line)...)
				}
			}
		}
	}
	return file.FilterExtractResultsByProbability(results, cutoffProbability), nil
}

func extractFromCommitIterator(cIter object.CommitIter, cutoffProbability float32) ([]file.ExtractResult, error) {
	var prev *object.Commit
	var results []file.ExtractResult
	err := cIter.ForEach(func(c *object.Commit) error {
		if prev == nil {
			prev = c
			return nil
		}
		commitResults, err := inspectCommit(c, prev, 0.5)
		if err != nil {
			return err
		}
		results = append(results, commitResults...)
		prev = c
		return nil
	})
	if err != nil {
		return nil, err
	}

	return file.FilterDuplicateExtractResults(results), nil
}

func ExtractGit(repoUrl string) error {
	repo, err := gitgo.PlainOpen(repoUrl)
	if err != nil {
		return err
	}

	cIter, err := repo.Log(&gitgo.LogOptions{})
	if err != nil {
		return err
	}

	results, err := extractFromCommitIterator(cIter, 0.5)
	if err != nil {
		return err
	}

	for _, result := range results {
		fmt.Println(result.String())
	}

	return nil
}

func ExtractGitFromCommit(repoUrl string, commitHash string, cutoffProbability float32) error {
	repo, err := gitgo.PlainOpen(repoUrl)
	if err != nil {
		return err
	}

	_, err = repo.CommitObject(plumbing.NewHash(commitHash))
	if err != nil {
		return err
	}

	cIter, err := repo.Log(&gitgo.LogOptions{From: plumbing.NewHash(commitHash)})
	if err != nil {
		return err
	}

	results, err := extractFromCommitIterator(cIter, cutoffProbability)
	if err != nil {
		return err
	}

	for _, result := range results {
		fmt.Println(result.String())
	}

	return nil
}
