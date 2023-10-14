package postgres

import (
	"context"

	"github.com/tedyst/licenta/scanner"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

type postgresScanner struct {
	db *pgx.Conn
}

type postgresScanResult struct {
	severity scanner.Severity
	message  string
	detail   string
}

func (result *postgresScanResult) String() string {
	return result.message
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
	_, err := sc.db.Query(ctx, "SELECT * FROM information_schema.role_table_grants;")
	if err != nil {
		return errors.Wrap(err, "could not see table information_schema.role_table_grants")
	}
	_, err = sc.db.Query(ctx, "SELECT * FROM pg_catalog.pg_role;")
	if err != nil {
		return errors.Wrap(err, "could not see table pg_catalog.pg_role")
	}
	_, err = sc.db.Query(ctx, "SELECT * FROM pg_catalog.pg_user;")
	if err != nil {
		return errors.Wrap(err, "could not see table pg_catalog.pg_user")
	}
	_, err = sc.db.Query(ctx, "SELECT * FROM pg_settings;")
	if err != nil {
		return errors.Wrap(err, "could not see table pg_settings")
	}
	_, err = sc.db.Query(ctx, "SELECT * FROM pg_file_settings;")
	if err != nil {
		return errors.Wrap(err, "could not see table pg_file_settings")
	}

	return nil
}

func (sc *postgresScanner) GetUsers(ctx context.Context) ([]scanner.User, error) {
	rows, err := sc.db.Query(ctx, "SELECT rolsuper, rolname, rolpassword FROM pg_catalog.pg_authid WHERE rolcanlogin=true;")

}

var _ scanner.Scanner = (*postgresScanner)(nil)
var _ scanner.ScanResult = (*postgresScanResult)(nil)
