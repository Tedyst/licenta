package redis

import (
	"context"

	"github.com/tedyst/licenta/scanner"
)

var scanConfigLines = map[string]struct {
	value      func(string) bool
	diagnostic redisScanResult
}{
	"timeout": {
		value: func(s string) bool { return s == "0" },
		diagnostic: redisScanResult{
			severity: scanner.SEVERITY_WARNING,
			message:  "timeout is set to 0.",
			detail:   "timeout is set to 0. This can cause the server to hang indefinitely",
		},
	},
	"tls-cert-file": {
		value: func(s string) bool { return s == "" },
		diagnostic: redisScanResult{
			severity: scanner.SEVERITY_HIGH,
			message:  "tls-cert-file is empty. TLS is not configured.",
			detail:   "tls-cert-file is empty. TLS is not configured.",
		},
	},
	"ignore-warnings": {
		value: func(s string) bool { return s != "" },
		diagnostic: redisScanResult{
			severity: scanner.SEVERITY_WARNING,
			message:  "ignore-warnings is not empty. This can hide important warnings.",
			detail:   "ignore-warnings is empty. This can hide important warnings.",
		},
	},
	"enable-debug-command": {
		value: func(s string) bool { return s != "no" },
		diagnostic: redisScanResult{
			severity: scanner.SEVERITY_HIGH,
			message:  "enable-debug-command is not set to no.",
			detail:   "enable-debug-command is not set to no.",
		},
	},
	"requirepass": {
		value: func(s string) bool { return s != "" },
		diagnostic: redisScanResult{
			severity: scanner.SEVERITY_MEDIUM,
			message:  "requirepass is set. Please consider using ACLs instead.",
			detail:   "requirepass is set. Please consider using ACLs instead.",
		},
	},
	"aclfile": {
		value: func(s string) bool { return s == "" },
		diagnostic: redisScanResult{
			severity: scanner.SEVERITY_HIGH,
			message:  "aclfile is not set. Please consider using ACLs.",
			detail:   "aclfile is not set. Please consider using ACLs.",
		},
	},
}

func (sc *redisScanner) ScanConfig(ctx context.Context) ([]scanner.ScanResult, error) {
	var results = make([]scanner.ScanResult, 0)
	config := sc.db.Do(ctx, "CONFIG", "GET", "*").Val().(map[interface{}]interface{})

	for key, value := range config {
		if line, ok := scanConfigLines[key.(string)]; ok {
			if line.value(value.(string)) {
				results = append(results, &line.diagnostic)
			}
		}
	}

	return results, nil
}
