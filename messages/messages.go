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
	ScanType       scanType
	PostgresScanID int
	ProjectID      int
}

type Exchange interface {
	PublishSendScanToWorkerMessage(ctx context.Context, worker models.Worker, message SendScanToWorkerMessage) error
	ReceiveSendScanToWorkerMessage(ctx context.Context, worker models.Worker) (SendScanToWorkerMessage, bool, error)
}

func GetPostgresScanMessage(postgresScanID int, projectID int) SendScanToWorkerMessage {
	return SendScanToWorkerMessage{
		ScanType:       PostgresScan,
		PostgresScanID: postgresScanID,
		ProjectID:      projectID,
	}
}
