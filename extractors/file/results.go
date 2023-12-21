package file

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

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
