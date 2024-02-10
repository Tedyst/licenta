package file

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"log/slog"

	"errors"
)

const maxMatchesPerLine = 100
const keepPreviousLines = 5

type secretType struct {
	regex          *regexp.Regexp
	probability    func(line string, match string) float64
	name           string
	postProcessing func(string) (string, error)
}

type FileScanner struct {
	wordsReduceProbability    map[string]struct{}
	wordsIncreaseProbability  map[string]struct{}
	passwordsCompletelyIgnore map[string]struct{}
	usernamesCompletelyIgnore map[string]struct{}

	options options

	secretTypes []secretType

	initiated bool
}

func stringsToMap(str []string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, s := range str {
		m[s] = struct{}{}
	}
	return m
}

func NewScanner(opts ...Option) (*FileScanner, error) {
	o, err := makeOptions(opts...)
	if err != nil {
		return nil, err
	}

	fs := &FileScanner{
		wordsReduceProbability:    stringsToMap(o.wordsReduceProbability),
		wordsIncreaseProbability:  stringsToMap(o.wordsIncreaseProbability),
		passwordsCompletelyIgnore: stringsToMap(o.passwordsCompletelyIgnore),
		usernamesCompletelyIgnore: stringsToMap(o.usernamesCompletelyIgnore),

		options: *o,

		secretTypes: []secretType{},

		initiated: true,
	}

	fs.secretTypes = fs.getSecretTypes()
	return fs, nil
}

func (fs *FileScanner) createResult(ctx context.Context, secretType secretType, match string, fileName string, lineNumber int, line string, previousLines string) (ExtractResult, bool, error) {
	result := ExtractResult{
		Name:          secretType.name,
		Line:          line,
		LineNumber:    lineNumber,
		PreviousLines: previousLines,
		FileName:      fileName,
	}

	if secretType.postProcessing != nil {
		postProcessed, err := secretType.postProcessing(match)
		if err != nil {
			return ExtractResult{}, false, nil
		}
		match = postProcessed
	}

	if len(secretType.regex.SubexpNames()) > 0 {
		for i, name := range secretType.regex.SubexpNames() {
			if i != 0 && name != "" {
				switch name {
				case "username":
					result.Username = secretType.regex.FindStringSubmatch(line)[i]
				case "password":
					result.Password = secretType.regex.FindStringSubmatch(line)[i]
				}
			}
		}
	}

	result.Match = strings.TrimSpace(match)

	if secretType.probability != nil {
		if result.Password != "" {
			result.Probability = secretType.probability(result.Line, result.Password)
		} else {
			result.Probability = secretType.probability(result.Line, result.Match)
		}
	} else {
		result.Probability = 1.0
	}

	if result.Probability < fs.options.minimumProbability {
		slog.DebugContext(ctx, "Probability too low", "probability", result.Probability, "fileName", fileName, "lineNumber", lineNumber)
		return ExtractResult{}, false, nil
	}

	if result.Password != "" {
		if len(result.Password) < 4 {
			slog.DebugContext(ctx, "Password too short", "password", result.Password, "fileName", fileName, "lineNumber", lineNumber)
			return ExtractResult{}, false, nil
		}
		if _, ok := fs.passwordsCompletelyIgnore[result.Password]; ok {
			slog.DebugContext(ctx, "Password completely ignored", "password", result.Password, "fileName", fileName, "lineNumber", lineNumber)
			return ExtractResult{}, false, nil
		}
	}

	if result.Username != "" {
		if _, ok := fs.usernamesCompletelyIgnore[result.Username]; ok {
			slog.DebugContext(ctx, "Username completely ignored", "username", result.Username, "fileName", fileName, "lineNumber", lineNumber)
			return ExtractResult{}, false, nil
		}
	}

	return result, true, nil
}

func (fs *FileScanner) ExtractFromLine(ctx context.Context, fileName string, lineNumber int, line string, previousLines string) ([]ExtractResult, error) {
	if !fs.initiated {
		return nil, errors.New("FileScanner not initiated")
	}

	var results []ExtractResult

	for _, secretType := range fs.secretTypes {
		for _, match := range secretType.regex.FindAllString(line, maxMatchesPerLine) {
			result, ok, err := fs.createResult(ctx, secretType, match, fileName, lineNumber, line, previousLines)
			if err != nil {
				return nil, err
			}

			if !ok {
				continue
			}

			results = append(results, result)
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("ExtractFromLine: context is done: %w", ctx.Err())
			default:
			}
		}
	}

	return results, nil
}

func (fs *FileScanner) ExtractFromReader(ctx context.Context, fileName string, rd io.Reader) ([]ExtractResult, error) {
	if !fs.initiated {
		return nil, errors.New("FileScanner not initiated")
	}

	var results []ExtractResult
	var lineNumber int

	slog.DebugContext(ctx, "Extracting from reader", "fileName", fileName)

	previousLines := []string{}
	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		if lineNumber > keepPreviousLines {
			previousLines = previousLines[1:]
		}
		line := scanner.Text()
		lineNumber++

		extracted, err := fs.ExtractFromLine(ctx, fileName, lineNumber, line, strings.Join(previousLines, "\n"))
		if err != nil {
			return nil, err
		}

		results = append(results, extracted...)
		previousLines = append(previousLines, line)

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("ExtractFromReader: context is done: %w", ctx.Err())
		default:
		}
	}
	return results, nil
}
