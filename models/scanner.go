package models

import "github.com/tedyst/licenta/db/queries"

type PostgresDatabases = queries.PostgresDatabase
type PostgresScan = queries.PostgresScan

type Scan = queries.Scan
type ScanResult = queries.ScanResult
type ScanBruteforceResult = queries.ScanBruteforceResult

const (
	SCAN_NOT_STARTED int32 = iota
	SCAN_RUNNING
	SCAN_FINISHED
	SCAN_QUEUED
	SCAN_CHECKING_PUBLIC_ACCESS
)
