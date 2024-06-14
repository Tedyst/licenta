package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/tedyst/licenta/tasks"
)

type taskSender struct {
	dockerScannerTaskSender
	emailSenderTaskSender
	gitScannerTaskSender
	natsScannerTaskSender
	nvdScannerTaskSender
}

func NewTaskSender(conn *nats.Conn) tasks.TaskRunner {
	return &taskSender{
		dockerScannerTaskSender: dockerScannerTaskSender{
			conn: conn,
		},
		emailSenderTaskSender: emailSenderTaskSender{
			conn: conn,
		},
		gitScannerTaskSender: gitScannerTaskSender{
			conn: conn,
		},
		natsScannerTaskSender: natsScannerTaskSender{
			conn: conn,
		},
		nvdScannerTaskSender: nvdScannerTaskSender{
			conn: conn,
		},
	}
}

var _ tasks.TaskRunner = &taskSender{}
