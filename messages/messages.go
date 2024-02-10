package messages

import (
	"context"

	"github.com/tedyst/licenta/db/queries"
)

type scanType string

const (
	PostgresScan scanType = "postgres_scan"
)

type SendScanToWorkerMessage struct {
	ScanID int64 `json:"scan_id"`
}

type Exchange interface {
	PublishSendScanToWorkerMessage(ctx context.Context, worker *queries.Worker, message SendScanToWorkerMessage) error
	ReceiveSendScanToWorkerMessage(ctx context.Context, worker *queries.Worker) (SendScanToWorkerMessage, bool, error)
}

func GetStartScanMessage(scan *queries.Scan) SendScanToWorkerMessage {
	return SendScanToWorkerMessage{
		ScanID: scan.ID,
	}
}
