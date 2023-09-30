package file

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"strings"

	"golang.org/x/exp/slog"
)

type secretType struct {
	regex          *regexp.Regexp
	probability    func(line string, match string) float64
	name           string
	postProcessing func(string) (string, error)
}

type ExtractResult struct {
	Name        string
	Line        string
	LineNumber  int
	Match       string
	Probability float64
	Username    string
	Password    string
	FileName    string
}

func (e ExtractResult) Hash() string {
	hasher := sha256.New()
	hasher.Write([]byte(e.Name))
	hasher.Write([]byte(e.Username))
	hasher.Write([]byte(e.Password))
	hasher.Write([]byte(e.FileName))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (e ExtractResult) String() string {
	return fmt.Sprintf("ExtractResult{Name: %s, Line: %s, LineNumber: %d, Match: %s, Probability: %f, Username: %s, Password: %s, FileName: %s}", e.Name, e.Line, e.LineNumber, e.Match, e.Probability, e.Username, e.Password, e.FileName)
}

func ExtractFromLine(ctx context.Context, fileName string, lineNumber int, line string, opts ...Option) ([]ExtractResult, error) {
	slog.DebugContext(ctx, "Extracting from line", "fileName", fileName, "lineNumber", lineNumber, "line", line)

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
			slog.DebugContext(ctx, "Found match", "match", match)
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
					slog.DebugContext(ctx, "Password completely ignored: %s", result.Password, "fileName", fileName, "lineNumber", lineNumber)
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
	}
	return results, nil
}
