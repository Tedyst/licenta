package scanner

import (
	"context"
)

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
	GetVersion(ctx context.Context) (string, error)
}

type ScanResult interface {
	Severity() Severity
	Detail() string
}

type User interface {
	GetUsername() (string, error)
	HasPassword() (bool, error)
	VerifyPassword(string) (bool, error)
	GetRawPassword() (string, bool, error)
	IsPrivileged() (bool, error)
	GetHashedPassword() (string, error)
}
