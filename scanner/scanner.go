package scanner

import "context"

type Severity = int

const (
	SEVERITY_WARNING Severity = iota
	SEVERITY_MEDIUM
	SEVERITY_HIGH
)

type Scanner interface {
	Ping(context.Context) error
	CheckPermissions(context.Context) error
	ScanConfig(ctx context.Context) ([]ScanResult, error)
	GetUsers(ctx context.Context) ([]User, error)
}

type ScanResult interface {
	String() string
	Severity() Severity
	Detail() string
}

type User interface {
	VerifyPassword(string) (bool, error)
}
