package messages

import (
	"context"

	"github.com/tedyst/licenta/models"
)

type scanType string

const (
	PostgresScan scanType = "postgres_scan"
)

type SendScanToWorkerMessage struct {
	ScanType  scanType `json:"scan_type"`
	ScanID    int64    `json:"scan_id"`
	ProjectID int64    `json:"project_id"`
}

type Exchange interface {
	PublishSendScanToWorkerMessage(ctx context.Context, worker *models.Worker, message *SendScanToWorkerMessage) error
	ReceiveSendScanToWorkerMessage(ctx context.Context, worker *models.Worker) (SendScanToWorkerMessage, bool, error)
}

func GetStartScanMessage(scan *models.Scan) SendScanToWorkerMessage {
	return SendScanToWorkerMessage{
		ScanType:  PostgresScan,
		ScanID:    scan.ID,
		ProjectID: scan.ProjectID,
	}
}
