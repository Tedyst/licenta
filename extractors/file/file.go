package file

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"regexp"
	"strings"
)

type secretType struct {
	regex          *regexp.Regexp
	severity       int
	probability    func(line string, match string) float32
	name           string
	postProcessing func(string) (string, error)
}

type ExtractResult struct {
	Name        string
	Line        string
	LineNumber  int
	Match       string
	Severity    int
	Probability float32
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
	return fmt.Sprintf("ExtractResult{Name: %s, Line: %s, LineNumber: %d, Match: %s, Severity: %d, Probability: %f, Username: %s, Password: %s, FileName: %s}", e.Name, e.Line, e.LineNumber, e.Match, e.Severity, e.Probability, e.Username, e.Password, e.FileName)
}

const entropyThreshold = 50

func shannonEntropy(value string) (bits int) {
	frq := make(map[rune]float64)

	//get frequency of characters
	for _, i := range value {
		frq[i]++
	}

	var sum float64

	for _, v := range frq {
		f := v / float64(len(value))
		sum += f * math.Log2(f)
	}

	bits = int(math.Ceil(sum*-1)) * len(value)
	return
}

func isASCII(s []byte) bool {
	for i := range s {
		if s[i] < 32 || s[i] > 126 {
			return false
		}
	}
	return true
}

var wordsReduceProbability = []string{
	"password",
	"error",
	"username",
	"login",
	"secret",
	"token",
	"key",
	"api",
	"access",
	"private",
	"public",
	"protected",
	"admin",
	"root",
	"user",
	"args",
	"null",
	"hash",
	"string",
	"request",
}

var wordsIncreaseProbability = []string{
	"postgres",
	"psql",
	"mongo",
	"mysql",
	"mariadb",
	"redis",
	"rabbitmq",
}

func calculateProbabilityCommon(line string, match string) float32 {
	entropy := shannonEntropy(match)
	probability := math.Min(float64(entropy)/float64(entropyThreshold), 1.0)
	for _, word := range wordsReduceProbability {
		if strings.Contains(strings.ToLower(match), strings.ToLower(word)) {
			probability /= 2
		}
	}
	for _, word := range wordsIncreaseProbability {
		if strings.Contains(strings.ToLower(line), strings.ToLower(word)) {
			probability *= 2
		}
	}
	return float32(math.Min(float64(probability), 1.0))
}

var secretTypes = []secretType{
	{
		regex:       regexp.MustCompile(`(?i)(_|[a-zA-Z])*(password|passwd|pwd|pass)(_|[a-zA-Z])* *(=|:|:=) *(?P<password>[a-zA-Z_\-\.]+)`),
		severity:    1,
		probability: calculateProbabilityCommon,
		name:        "Generic Password",
	},
	{
		regex:    regexp.MustCompile(`(?i)postgres:\/\/(?P<username>[^:]+)( *)(=|:)( *)(?P<password>[^@]+)@`),
		severity: 1,
		name:     "Postgres Connection String",
	},
	{
		regex:    regexp.MustCompile(`(?i)mongodb:\/\/(?P<username>[^:]+)( *)(=|:)( *)(?P<password>[^@]+)@`),
		severity: 1,
		name:     "MongoDB Connection String",
	},
	{
		regex:    regexp.MustCompile(`(?i)(?P<username>(?:\b|_)([a-zA-Z0-9_]*(?:api|key|token)[a-zA-Z0-9_]*))(=|:)("|')?(?P<password>[a-zA-Z0-9_\-\.]*)(?:\b|_)("|')?`),
		severity: 1,
		name:     "Generic Environment Variable",
	},
	{
		regex:       regexp.MustCompile(`(?:[A-Za-z0-9+\/]{4})*(?:[A-Za-z0-9+\/]{4}|[A-Za-z0-9+\/]{3}=|[A-Za-z0-9+\/]{2}={2})`),
		severity:    1,
		probability: calculateProbabilityCommon,
		name:        "Generic Base64",
		postProcessing: func(match string) (string, error) {
			if len(match) < 7 {
				return "", fmt.Errorf("too short")
			}
			str, err := base64.StdEncoding.DecodeString(match)
			if err != nil {
				return "", err
			}
			if !isASCII(str) {
				return "", fmt.Errorf("not ascii")
			}
			return string(str), nil
		},
	},
}

func ExtractFromLine(fileName string, lineNumber int, line string) []ExtractResult {
	var results []ExtractResult

	for _, secretType := range secretTypes {
		for _, match := range secretType.regex.FindAllString(line, 100) {
			result := ExtractResult{
				Name:       secretType.name,
				Line:       strings.TrimSpace(line),
				LineNumber: lineNumber,
				Match:      match,
				Severity:   secretType.severity,
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

			results = append(results, result)
		}
	}
	return results
}

func ExtractFromReader(fileName string, rd io.Reader) ([]ExtractResult, error) {
	var results []ExtractResult
	var lineNumber int

	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++

		extracted := ExtractFromLine(fileName, lineNumber, line)
		results = append(results, extracted...)
	}
	return results, nil
}

func FilterDuplicateExtractResults(originalResults []ExtractResult) []ExtractResult {
	var results map[string]ExtractResult = make(map[string]ExtractResult)
	for _, result := range originalResults {
		if result.Probability > results[result.Hash()].Probability {
			results[result.Hash()] = result
		}
	}
	var filteredResults []ExtractResult
	for _, result := range results {
		filteredResults = append(filteredResults, result)
	}
	return filteredResults
}

func FilterExtractResultsByProbability(originalResults []ExtractResult, probability float32) []ExtractResult {
	var results []ExtractResult
	for _, result := range originalResults {
		if result.Probability >= probability {
			results = append(results, result)
		}
	}
	return results
}
