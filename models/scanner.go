package models

import "github.com/tedyst/licenta/db/queries"

type PostgresDatabases = queries.PostgresDatabase
type PostgresScan = queries.PostgresScan
type PostgresScanResult = queries.PostgresScanResult
type PostgresScanBruteforceResult = queries.PostgresScanBruteforceResult

const (
	SCAN_NOT_STARTED int32 = iota
	SCAN_RUNNING
	SCAN_FINISHED
)
