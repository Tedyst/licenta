package file

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"math"
	"regexp"
	"strings"
)

const defaultProbabilityDecreaseMultiplier = 0.7
const defaultProbabilityIncreaseMultiplier = 2.0
const defaultEntropyThresholdMidpoint = 40
const defaultLogisticGrowthRate = 0.2

func (fs *FileScanner) getSecretTypes() []secretType {
	calculateProbabilityCommonWithMultiplier := func(multiplier float64) func(string, string) float64 {
		return func(line string, match string) float64 {
			entropy := shannonEntropy(match)
			probability := 1.0 / (1.0 + math.Exp(-fs.options.logisticGrowthRate*float64(entropy-fs.options.entropyThresholdMidpoint)))

			scanner := bufio.NewScanner(strings.NewReader(strings.ToLower(match)))
			scanner.Split(bufio.ScanWords)

			for scanner.Scan() {
				lower := strings.ToLower(scanner.Text())
				if _, ok := fs.wordsIncreaseProbability[lower]; ok {
					probability *= fs.options.probabilityDecreaseMultiplier
				}
				if _, ok := fs.wordsIncreaseProbability[lower]; ok {
					probability *= fs.options.probabilityIncreaseMultiplier
				}
			}
			return math.Min(probability*multiplier, 1.0)
		}
	}

	return []secretType{
		{
			regex:       regexp.MustCompile(`(?i)(_|[a-zA-Z])*(password|passwd|pwd|pass)(_|[a-zA-Z])* *(=|:|:=) *(?P<password>[a-zA-Z_\-\.]+)`),
			probability: calculateProbabilityCommonWithMultiplier(1),
			name:        "Generic Password",
		},
		{
			regex: regexp.MustCompile(`(?i)postgres:\/\/(?P<username>[^:]+)( *)(=|:)( *)(?P<password>[^@]+)@`),
			name:  "Postgres Connection String",
		},
		{
			regex:       regexp.MustCompile(`(?i)(?P<username>[a-zA-Z0-9\-\.]+)(=|:)(?P<password>[a-zA-Z0-9\-\.]+)@`),
			name:        "Generic Connection String",
			probability: calculateProbabilityCommonWithMultiplier(1),
		},
		{
			regex: regexp.MustCompile(`(?i)mongodb:\/\/(?P<username>[^:]+)( *)(=|:)( *)(?P<password>[^@]+)@`),
			name:  "MongoDB Connection String",
		},
		{
			regex: regexp.MustCompile(`(?i)(?P<username>(?:\b|_)([a-zA-Z0-9_]*(?:api|key|token)[a-zA-Z0-9_]*))(=|:)("|')?(?P<password>[a-zA-Z0-9_\-\.]+)(?:\b|_)("|')?`),
			name:  "Generic Environment Variable",
		},
		{
			regex:       regexp.MustCompile(`(?:[A-Za-z0-9+\/]{4})*(?:[A-Za-z0-9+\/]{4}|[A-Za-z0-9+\/]{3}=|[A-Za-z0-9+\/]{2}={2})`),
			probability: calculateProbabilityCommonWithMultiplier(2),
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
}

var defaultWordsReduceProbability = []string{
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
	".txt",
	".cfg",
}

var defaultPasswordsCompletelyIgnore = []string{
	"password",
	"string",
	"request",
	"value",
	"example",
	"lambda",
	"true",
	"false",
	"none",
	"function",
	"no",
	"yes",
	"amd64",
	"arm64",
	"arm",
	"linux",
	"darwin",
	"windows",
	"amd",
	"macos",
	"i386",
	"android",
	"ios",
	"example",
	"keyid",
	"kubernetescluster",
	"runtime",
	"name",
	"secret_access_key",
	"setting",
	"api_key",
	"api_secret",
	"key_id",
	"key_secret",
	"always",
	"1024",
	"2048",
	"4096",
	"token",
	"hash",
	"suspend",
	"caller",
	"getvalidator",
	"fields",
	"hibernate",
	"poweroff",
	"reboot",
	"author",
	"sql.NullString",
}

var defaultUsernamesCompletelyIgnore = []string{
	"i18nKey",
	"assetkey",
}

var defaultWordsIncreaseProbability = []string{
	"database",
	"db",
	"postgres",
	"psql",
	"mongo",
	"mysql",
	"mariadb",
	"redis",
	"rabbitmq",
}
