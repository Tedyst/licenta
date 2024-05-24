package scanner

import (
	"context"
	"errors"

	"github.com/tedyst/licenta/nvd"
)

type Severity = int

const (
	SEVERITY_INFORMATIONAL Severity = iota
	SEVERITY_WARNING
	SEVERITY_MEDIUM
	SEVERITY_HIGH
)

var (
	ErrPingNotSupported             = errors.New("ping not supported")
	ErrCheckPermissionsNotSupported = errors.New("check permissions not supported")
	ErrScanConfigNotSupported       = errors.New("scan config not supported")
	ErrGetUsersNotSupported         = errors.New("get users not supported")
	ErrVersionNotSupported          = errors.New("version not supported")
)

type Scanner interface {
	GetScannerName() string
	GetScannerID() int32

	GetNvdProductType() nvd.Product
	ShouldNotBePublic() bool

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
