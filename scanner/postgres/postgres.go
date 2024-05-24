package postgres

import (
	"context"
	"fmt"

	"github.com/tedyst/licenta/nvd"
	"github.com/tedyst/licenta/scanner"

	"github.com/jackc/pgx/v5"
)

type postgresScanner struct {
	db *pgx.Conn
}

var _ scanner.Scanner = (*postgresScanner)(nil)
var _ scanner.ScanResult = (*postgresScanResult)(nil)

type postgresScanResult struct {
	severity scanner.Severity
	message  string
	detail   string
}

func (result *postgresScanResult) Severity() scanner.Severity {
	return result.severity
}

func (result *postgresScanResult) Detail() string {
	return result.detail
}

func (sc *postgresScanner) Ping(ctx context.Context) error {
	return sc.db.Ping(ctx)
}

func (sc *postgresScanner) CheckPermissions(ctx context.Context) error {
	row, err := sc.db.Query(ctx, "SELECT * FROM information_schema.role_table_grants;")
	if err != nil {
		return fmt.Errorf("could not see table information_schema.role_table_grants: %w", err)
	}
	row.Close()
	row, err = sc.db.Query(ctx, "SELECT * FROM pg_catalog.pg_roles;")
	if err != nil {
		return fmt.Errorf("could not see table pg_catalog.pg_roles: %w", err)
	}
	row.Close()
	row, err = sc.db.Query(ctx, "SELECT * FROM pg_catalog.pg_user;")
	if err != nil {
		return fmt.Errorf("could not see table pg_catalog.pg_user: %w", err)
	}
	row.Close()
	row, err = sc.db.Query(ctx, "SELECT * FROM pg_settings;")
	if err != nil {
		return fmt.Errorf("could not see table pg_settings: %w", err)
	}
	row.Close()
	row, err = sc.db.Query(ctx, "SELECT * FROM pg_file_settings;")
	if err != nil {
		return fmt.Errorf("could not see table pg_file_settings: %w", err)
	}
	row.Close()

	return nil
}

func (sc *postgresScanner) GetNvdProductType() nvd.Product {
	return nvd.POSTGRESQL
}

func (sc *postgresScanner) ShouldNotBePublic() bool {
	return true
}

func (*postgresScanner) GetScannerID() int32 {
	return 0
}

func GetScannerID() int32 {
	return 0
}

func (*postgresScanner) GetScannerName() string {
	return "PostgreSQL"
}

func NewScanner(ctx context.Context, db *pgx.Conn) (scanner.Scanner, error) {
	sc := &postgresScanner{
		db: db,
	}

	return sc, nil
}
