package postgres

import "context"

func (sc *postgresScanner) GetVersion(ctx context.Context) (string, error) {
	var version string
	err := sc.db.QueryRow(ctx, "SELECT version();").Scan(&version)
	if err != nil {
		return "", err
	}
	return version, nil
}
