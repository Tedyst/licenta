package postgres

import "context"

func (sc *postgresScanner) GetVersion(ctx context.Context) (string, error) {
	var version string
	err := sc.db.QueryRow(ctx, "SELECT substring(version() from 'PostgreSQL ([0-9]+\\.[0-9]+)') as version_number;").Scan(&version)
	if err != nil {
		return "", err
	}

	return version, nil
}
