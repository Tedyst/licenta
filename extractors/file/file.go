package file

import (
	"bufio"
	"context"
	"io"
	"regexp"
	"strings"

	"log/slog"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/pkg/errors"
)

type secretType struct {
	regex          *regexp.Regexp
	probability    func(line string, match string) float64
	name           string
	postProcessing func(string) (string, error)
}

type fileScanner struct {
	wordsReduceProbabilityTrie    mapset.Set[string]
	wordsIncreaseProbabilityTrie  mapset.Set[string]
	passwordsCompletelyIgnoreTrie mapset.Set[string]
	usernamesCompletelyIgnoreTrie mapset.Set[string]

	options options

	secretTypes []secretType
}

func ExtractFromLine(ctx context.Context, fileName string, lineNumber int, line string, opts ...Option) ([]ExtractResult, error) {
	o, err := makeOptions(opts...)
	if err != nil {
		return nil, err
	}

	var results []ExtractResult
	wordsReduceProbabilityTrie := GetTrie(o.wordsReduceProbability)
	wordsIncreaseProbabilityTrie := GetTrie(o.wordsIncreaseProbability)
	passwordsCompletelyIgnoreTrie := GetTrie(o.passwordsCompletelyIgnore)
	usernamesCompletelyIgnoreTrie := GetTrie(o.usernamesCompletelyIgnore)

	secretTypes := getSecretTypes(
		wordsReduceProbabilityTrie,
		wordsIncreaseProbabilityTrie,
		o.logisticGrowthRate,
		o.entropyThresholdMidpoint,
		o.probabilityDecreaseMultiplier,
		o.probabilityIncreaseMultiplier,
	)

	for _, secretType := range secretTypes {
		for _, match := range secretType.regex.FindAllString(line, 100) {
			result := ExtractResult{
				Name:       secretType.name,
				Line:       strings.TrimSpace(line),
				LineNumber: lineNumber,
				Match:      match,
				FileName:   fileName,
			}
			if secretType.postProcessing != nil {
				postProcessed, err := secretType.postProcessing(match)
				if err != nil {
					continue
				}
				result.Match = postProcessed
			}
			result.Match = strings.TrimSpace(result.Match)
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

			if result.Username == "" && result.Password == "" {
				continue
			}

			if secretType.probability != nil {
				if result.Password != "" {
					result.Probability = secretType.probability(result.Line, result.Password)
				} else {
					result.Probability = secretType.probability(result.Line, result.Match)
				}
			} else {
				result.Probability = 1.0
			}

			if result.Password != "" {
				if len(result.Password) < 4 {
					slog.DebugContext(ctx, "Password too short", "password", result.Password, "fileName", fileName, "lineNumber", lineNumber)
					continue
				}
				if passwordsCompletelyIgnoreTrie.Get(strings.ToLower(result.Password)) != nil {
					slog.DebugContext(ctx, "Password completely ignored", "password", result.Password, "fileName", fileName, "lineNumber", lineNumber)
					continue
				}
			}
			if result.Username != "" {
				if usernamesCompletelyIgnoreTrie.Get(strings.ToLower(result.Username)) != nil {
					slog.DebugContext(ctx, "Username completely ignored", "username", result.Username, "fileName", fileName, "lineNumber", lineNumber)
					continue
				}
			}

			results = append(results, result)
			select {
			case <-ctx.Done():
				return nil, errors.Wrap(ctx.Err(), "ExtractFromLine")
			default:
			}
		}
	}
	return results, nil
}

func ExtractFromReader(ctx context.Context, fileName string, rd io.Reader, opts ...Option) ([]ExtractResult, error) {
	var results []ExtractResult
	var lineNumber int

	slog.DebugContext(ctx, "Extracting from reader", "fileName", fileName)

	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++

		extracted, err := ExtractFromLine(ctx, fileName, lineNumber, line, opts...)
		if err != nil {
			return nil, err
		}
		results = append(results, extracted...)
		select {
		case <-ctx.Done():
			return nil, errors.Wrap(ctx.Err(), "ExtractFromReader")
		default:
		}
	}
	return results, nil
}
