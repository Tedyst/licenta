package postgres

import (
	"context"

	"github.com/tedyst/licenta/nvd"
	"github.com/tedyst/licenta/scanner"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
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
		return errors.Wrap(err, "could not see table information_schema.role_table_grants")
	}
	row.Close()
	row, err = sc.db.Query(ctx, "SELECT * FROM pg_catalog.pg_roles;")
	if err != nil {
		return errors.Wrap(err, "could not see table pg_catalog.pg_roles")
	}
	row.Close()
	row, err = sc.db.Query(ctx, "SELECT * FROM pg_catalog.pg_user;")
	if err != nil {
		return errors.Wrap(err, "could not see table pg_catalog.pg_user")
	}
	row.Close()
	row, err = sc.db.Query(ctx, "SELECT * FROM pg_settings;")
	if err != nil {
		return errors.Wrap(err, "could not see table pg_settings")
	}
	row.Close()
	row, err = sc.db.Query(ctx, "SELECT * FROM pg_file_settings;")
	if err != nil {
		return errors.Wrap(err, "could not see table pg_file_settings")
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

func (*postgresScanner) GetScannerID() int64 {
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
