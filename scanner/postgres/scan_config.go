package postgres

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/tedyst/licenta/scanner"
)

var scanConfigLines = map[string]struct {
	value      func(string) bool
	diagnostic postgresScanResult
}{
	"fsync": {
		value: func(s string) bool { return s == "off" },
		diagnostic: postgresScanResult{
			severity: scanner.SEVERITY_WARNING,
			message:  "fsync is off. Loss of data in case of crash.",
			detail: "fsync=off is a dangerous setting. " +
				"Data loss is possible in case of a crash. " +
				"See https://www.postgresql.org/docs/current/runtime-config-wal.html#GUC-FSYNC for more details.",
		},
	},
	"data_directory_mode": {
		value: func(s string) bool { return s != "0700" },
		diagnostic: postgresScanResult{
			severity: scanner.SEVERITY_HIGH,
			message:  "data_directory_mode is not 0700.",
			detail:   "data_directory_mode is not 0700. This may allow other users to read your data.",
		},
	},
	"listen_addresses": {
		value: func(s string) bool { return s != "localhost" },
		diagnostic: postgresScanResult{
			severity: scanner.SEVERITY_MEDIUM,
			message:  "listen_addresses is not localhost.",
			detail:   "listen_addresses is not localhost. This may allow other users to connect to your database.",
		},
	},
	"full_page_writes": {
		value: func(s string) bool { return s == "off" },
		diagnostic: postgresScanResult{
			severity: scanner.SEVERITY_WARNING,
			message:  "full_page_writes is off. Loss of data in case of crash.",
			detail: "full_page_writes=off is a dangerous setting. " +
				"Data loss is possible in case of a crash. " +
				"See https://www.postgresql.org/docs/current/runtime-config-wal.html#GUC-FULL-PAGE-WRITES for more details.",
		},
	},
	"ssl": {
		value: func(s string) bool { return s == "off" },
		diagnostic: postgresScanResult{
			severity: scanner.SEVERITY_HIGH,
			message:  "ssl is off. Passwords are sent in clear text.",
			detail:   "ssl=off is a dangerous setting. Passwords are sent in clear text.",
		},
	},
	"idle_in_transaction_session_timeout": {
		value: func(s string) bool { return s != "0" },
		diagnostic: postgresScanResult{
			severity: scanner.SEVERITY_WARNING,
			message:  "idle_in_transaction_session_timeout is not 0.",
			detail:   "idle_in_transaction_session_timeout is not 0. This may cause a denial of service.",
		},
	},
	"ignore_invalid_pages": {
		value: func(s string) bool { return s == "off" },
		diagnostic: postgresScanResult{
			severity: scanner.SEVERITY_WARNING,
			message:  "ignore_invalid_pages is off.",
			detail:   "ignore_invalid_pages is off. This may cause error writing data to disk.",
		},
	},
	"local_preload_libraries": {
		value: func(s string) bool { return s != "" },
		diagnostic: postgresScanResult{
			severity: scanner.SEVERITY_HIGH,
			message:  "local_preload_libraries is not empty.",
			detail:   "local_preload_libraries is not empty. This may allow privilege escalation.",
		},
	},
	"log_connections": {
		value: func(s string) bool { return s == "off" },
		diagnostic: postgresScanResult{
			severity: scanner.SEVERITY_WARNING,
			message:  "log_connections is off.",
			detail:   "log_connections is off. This may make it harder to diagnose problems.",
		},
	},
	"log_disconnections": {
		value: func(s string) bool { return s == "off" },
		diagnostic: postgresScanResult{
			severity: scanner.SEVERITY_WARNING,
			message:  "log_disconnections is off.",
			detail:   "log_disconnections is off. This may make it harder to diagnose problems.",
		},
	},
	"log_file_mode": {
		value: func(s string) bool { return s != "0600" },
		diagnostic: postgresScanResult{
			severity: scanner.SEVERITY_HIGH,
			message:  "log_file_mode is not 0600.",
			detail:   "log_file_mode is not 0600. This may allow other users to read your logs.",
		},
	},
	"max_connections": {
		value: func(s string) bool {
			var maxConnections int
			_, err := fmt.Sscanf(s, "%d", &maxConnections)
			if err != nil {
				return false
			}
			return maxConnections > 1500
		},
		diagnostic: postgresScanResult{
			severity: scanner.SEVERITY_HIGH,
			message:  "max_connections is >1500.",
			detail:   "max_connections is >1500. This may cause a denial of service.",
		},
	},
	"password_encryption": {
		value: func(s string) bool { return s == "off" || s == "md5" },
		diagnostic: postgresScanResult{
			severity: scanner.SEVERITY_HIGH,
			message:  "password_encryption is off or set to md5",
			detail:   "password_encryption is off or set to md5. This may allow privilege escalation.",
		},
	},
	"syncronous_commit": {
		value: func(s string) bool { return s == "off" },
		diagnostic: postgresScanResult{
			severity: scanner.SEVERITY_WARNING,
			message:  "syncronous_commit is off.",
			detail:   "syncronous_commit is off. This may cause a loss of data in case of crash.",
		},
	},
	"TimeZone": {
		value: func(s string) bool { return s != "Etc/UTC" },
		diagnostic: postgresScanResult{
			severity: scanner.SEVERITY_WARNING,
			message:  "TimeZone is not UTC.",
			detail:   "TimeZone is not UTC. This may cause problems with timezones.",
		},
	},
}

func (sc *postgresScanner) ScanConfig(ctx context.Context) ([]scanner.ScanResult, error) {
	var results = make([]scanner.ScanResult, 0)

	rows, err := sc.db.Query(ctx, "SELECT name, setting FROM pg_settings;")
	if err != nil {
		return nil, errors.Wrap(err, "could not see table pg_settings")
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var setting string

		err := rows.Scan(&name, &setting)
		if err != nil {
			return nil, errors.Wrap(err, "could not scan row")
		}

		if line, ok := scanConfigLines[name]; ok {
			if line.value(setting) {
				diagnostic := line.diagnostic
				results = append(results, &diagnostic)
			}
		}
	}

	return results, nil
}
