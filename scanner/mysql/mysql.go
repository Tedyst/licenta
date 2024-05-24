package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/nvd"
	"github.com/tedyst/licenta/scanner"
)

type mysqlScanResult struct {
	severity scanner.Severity
	message  string
	detail   string
}

func (result *mysqlScanResult) Severity() scanner.Severity {
	return result.severity
}

func (result *mysqlScanResult) Detail() string {
	return result.detail
}

type mysqlScanner struct {
	db *sql.DB
}

func (sc *mysqlScanner) GetScannerName() string {
	return "MySQL"
}
func (sc *mysqlScanner) GetScannerID() int32 {
	return models.SCAN_MYSQL
}
func GetScannerID() int32 {
	return models.SCAN_MYSQL
}
func (sc *mysqlScanner) GetNvdProductType() nvd.Product {
	return nvd.MYSQL
}
func (sc *mysqlScanner) ShouldNotBePublic() bool {
	return true
}
func (sc *mysqlScanner) Ping(context.Context) error {
	return sc.db.Ping()
}
func (sc *mysqlScanner) CheckPermissions(ctx context.Context) error {
	_, err := sc.db.QueryContext(ctx, "SELECT * FROM mysql.user")
	if err != nil {
		return fmt.Errorf("could not see table mysql.user: %w", err)
	}

	_, err = sc.db.QueryContext(ctx, "SHOW VARIABLES")
	if err != nil {
		return fmt.Errorf("could not run SHOW VARIABLES: %w", err)
	}

	_, err = sc.db.QueryContext(ctx, "SELECT VERSION()")
	if err != nil {
		return fmt.Errorf("could not run SELECT VERSION(): %w", err)
	}

	return nil
}
func (sc *mysqlScanner) GetVersion(ctx context.Context) (string, error) {
	var version string
	err := sc.db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version)
	if err != nil {
		return "", fmt.Errorf("could not get version: %w", err)
	}

	return version, nil
}

func NewScanner(ctx context.Context, db *sql.DB) (scanner.Scanner, error) {
	sc := &mysqlScanner{
		db: db,
	}

	return sc, nil
}
