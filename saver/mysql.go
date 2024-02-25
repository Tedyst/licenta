package saver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/scanner/mysql"
)

type MysqlQuerier interface {
	BaseQuerier

	GetMysqlScanByScanID(ctx context.Context, scanID int64) (*queries.MysqlScan, error)
	GetMysqlDatabase(ctx context.Context, id int64) (*queries.GetMysqlDatabaseRow, error)
	UpdateMysqlVersion(ctx context.Context, params queries.UpdateMysqlVersionParams) error
}

func getMysqlConnectString(db *queries.MysqlDatabase) string {
	return fmt.Sprintf("%s:%s@%s:%d/%s", db.Username, db.Password, db.Host, db.Port, db.DatabaseName)
}

func NewMysqlSaver(ctx context.Context, baseQuerier BaseQuerier, bruteforceProvider bruteforce.BruteforceProvider, scan *queries.Scan) (Saver, error) {
	queries, ok := baseQuerier.(MysqlQuerier)
	if !ok {
		return nil, errors.Join(ErrSaverNotNeeded, fmt.Errorf("queries is not a MysqlQuerier"))
	}

	mysqlScan, err := queries.GetMysqlScanByScanID(ctx, scan.ID)
	if err != nil {
		return nil, errors.Join(ErrSaverNotNeeded, fmt.Errorf("could not get mysql scan: %w", err))
	}

	db, err := queries.GetMysqlDatabase(ctx, mysqlScan.DatabaseID)
	if err != nil {
		return nil, fmt.Errorf("could not get database: %w", err)
	}

	conn, err := sql.Open("mysql", getMysqlConnectString(&db.MysqlDatabase))
	if err != nil {
		return nil, fmt.Errorf("cannot create database connection: %w", err)
	}

	sc, err := mysql.NewScanner(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("could not create scanner: %w", err)
	}

	logger := slog.With(
		"scan", scan.ID,
		"Mysql_scan", mysqlScan.ID,
		"Mysql_database_id", mysqlScan.DatabaseID,
	)

	saver := &mysqlSaver{
		queries:    queries,
		baseSaver:  *createBaseSaver(queries, bruteforceProvider, logger, scan, sc),
		mysqlScan:  mysqlScan,
		database:   &db.MysqlDatabase,
		connection: conn,
	}
	saver.runAfterScan = saver.hookAfterScan
	return saver, nil
}

func (saver *mysqlSaver) hookAfterScan(ctx context.Context) error {
	version, err := saver.scanner.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("could not get version: %w", err)
	}
	if err := saver.queries.UpdateMysqlVersion(ctx, queries.UpdateMysqlVersionParams{
		ID:      saver.mysqlScan.DatabaseID,
		Version: sql.NullString{String: version, Valid: true},
	}); err != nil {
		return fmt.Errorf("could not update version: %w", err)
	}

	return saver.connection.Close()
}

type mysqlSaver struct {
	queries MysqlQuerier

	connection *sql.DB

	mysqlScan *queries.MysqlScan
	database  *queries.MysqlDatabase

	baseSaver
}

func init() {
	savers["mysql"] = NewMysqlSaver
}
