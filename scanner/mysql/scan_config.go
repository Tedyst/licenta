package mysql

import (
	"context"
	"fmt"
	"strconv"

	"github.com/tedyst/licenta/scanner"
)

var scanConfigLines = map[string]struct {
	value      func(string) bool
	diagnostic mysqlScanResult
}{
	"ssl_key": {
		value: func(s string) bool { return s == "" },
		diagnostic: mysqlScanResult{
			severity: scanner.SEVERITY_HIGH,
			message:  "ssl_key is empty. SSL is not configured.",
			detail: "ssl_key is empty. SSL is not configured. " +
				"See https://dev.mysql.com/doc/refman/8.0/en/server-system-variables.html#sysvar_ssl_key for more details.",
		},
	},
	"caching_sha2_password_digest_rounds": {
		value: func(s string) bool {
			t, err := strconv.Atoi(s)
			if err != nil {
				return false
			}
			return t < 5000
		},
		diagnostic: mysqlScanResult{
			severity: scanner.SEVERITY_MEDIUM,
			message:  "caching_sha2_password_digest_rounds is lower than 5000.",
			detail: "caching_sha2_password_digest_rounds is lower than 5000. " +
				"See https://dev.mysql.com/doc/refman/8.0/en/server-system-variables.html#sysvar_caching_sha2_password_digest_rounds for more details.",
		},
	},
	"debug": {
		value: func(s string) bool { return s != "d:t:O,/tmp/mysql.trace" },
		diagnostic: mysqlScanResult{
			severity: scanner.SEVERITY_WARNING,
			message:  "debug is enabled.",
			detail: "debug is enabled. " +
				"See https://dev.mysql.com/doc/refman/8.0/en/server-system-variables.html#sysvar_debug for more details.",
		},
	},
	"flush": {
		value: func(s string) bool { return s != "OFF" },
		diagnostic: mysqlScanResult{
			severity: scanner.SEVERITY_MEDIUM,
			message:  "flush is not OFF.",
			detail: "flush is not OFF. " +
				"See https://dev.mysql.com/doc/refman/8.0/en/server-system-variables.html#sysvar_flush for more details.",
		},
	},
}

func (sc *mysqlScanner) ScanConfig(ctx context.Context) ([]scanner.ScanResult, error) {
	var results = make([]scanner.ScanResult, 0)

	rows, err := sc.db.QueryContext(ctx, "SHOW VARIABLES")
	if err != nil {
		return nil, fmt.Errorf("could not run SHOW VARIABLES: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var setting string

		err := rows.Scan(&name, &setting)
		if err != nil {
			return nil, fmt.Errorf("could not scan row: %w", err)
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
