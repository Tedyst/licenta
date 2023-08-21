package file

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"regexp"
	"unicode"
)

type secretType struct {
	regex          *regexp.Regexp
	severity       int
	probability    float32
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
}

func isASCII(s []byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII || s[i] == ' ' || s[i] == '\n' || s[i] == '\t' || s[i] == '\r' {
			return false
		}
	}
	return true
}

var secretTypes = []secretType{
	{
		regex:       regexp.MustCompile(`(?i)( |_|[a-zA-Z])*(password|passwd|pwd|pass)( |_|[a-zA-Z])* *(=|:|:=) *(?P<password>[a-zA-Z_ -]*)`),
		severity:    1,
		probability: 1.0,
		name:        "Generic Password",
	},
	{
		regex:       regexp.MustCompile(`(?i)postgres:\/\/(?P<username>[^:]+)( *)(=|:)( *)(?P<password>[^@]+)@`),
		severity:    1,
		probability: 1.0,
		name:        "Postgres Connection String",
	},
	{
		regex:       regexp.MustCompile(`(?i)mongodb:\/\/(?P<username>[^:]+)( *)(=|:)( *)(?P<password>[^@]+)@`),
		severity:    1,
		probability: 1.0,
		name:        "MongoDB Connection String",
	},
	{
		regex:       regexp.MustCompile(`(?i)(?P<username>(?:\b|_)(.*(?:api|key|token).*))(=|:)("|')?(?P<password>.*)(?:\b|_)("|')?`),
		severity:    1,
		probability: 1.0,
		name:        "Generic Environment Variable",
	},
	{
		regex:       regexp.MustCompile(`(?:[A-Za-z0-9+\/]{4})*(?:[A-Za-z0-9+\/]{4}|[A-Za-z0-9+\/]{3}=|[A-Za-z0-9+\/]{2}={2})`),
		severity:    1,
		probability: 1.0,
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

var falsePositives = []*regexp.Regexp{}

func extractLine(lineNumber int, line string) []ExtractResult {
	var results []ExtractResult

	for _, secretType := range secretTypes {
		for _, match := range secretType.regex.FindAllString(line, 100) {
			result := ExtractResult{
				Name:        secretType.name,
				Line:        line,
				LineNumber:  lineNumber,
				Match:       match,
				Severity:    secretType.severity,
				Probability: secretType.probability,
			}
			if secretType.postProcessing != nil {
				postProcessed, err := secretType.postProcessing(match)
				if err != nil {
					continue
				}
				result.Match = postProcessed
			}
			for _, falsePositive := range falsePositives {
				if falsePositive.MatchString(result.Match) {
					result.Probability /= 2
				}
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

			results = append(results, result)
		}
	}
	return results
}

func ExtractFromReader(rd io.Reader) ([]ExtractResult, error) {
	var results []ExtractResult
	var lineNumber int

	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++

		extracted := extractLine(lineNumber, line)
		results = append(results, extracted...)
	}
	return results, nil
}
