package messages

import (
	"context"

	"github.com/tedyst/licenta/models"
)

type SendScanToWorkerMessage struct {
	PostgresScanID int
	ProjectID      int
}

type Exchange interface {
	PublishSendScanToWorkerMessage(ctx context.Context, worker models.Worker, postgresScanID int, projectID int) error
	ReceiveSendScanToWorkerMessage(ctx context.Context, worker models.Worker) (SendScanToWorkerMessage, bool, error)
}
