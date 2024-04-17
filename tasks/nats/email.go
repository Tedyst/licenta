package nats

import (
	"context"
	"log/slog"
	sync "sync"

	"github.com/nats-io/nats.go"
	"github.com/tedyst/licenta/tasks"
	"golang.org/x/sync/semaphore"
)

type emailSenderTaskSender struct {
	conn *nats.Conn
}

const sendEmailQueue = "send-email"

func NewEmailSenderTaskSender(conn *nats.Conn) *emailSenderTaskSender {
	return &emailSenderTaskSender{
		conn: conn,
	}
}

func (es *emailSenderTaskSender) SendEmail(ctx context.Context, address string, subject string, html string, text string) error {
	return publishMessage(ctx, es.conn, sendEmailQueue, &SendEmailMessage{
		Address: address,
		Subject: subject,
		Html:    html,
		Text:    text,
	}, 0)
}

type emailSenderTaskRunner struct {
	conn        *nats.Conn
	localRunner tasks.EmailTasksRunner
	semaphore   *semaphore.Weighted
}

func NewEmailSenderTaskRunner(conn *nats.Conn, localRunner tasks.EmailTasksRunner, concurrency int) *emailSenderTaskRunner {
	return &emailSenderTaskRunner{
		conn:        conn,
		localRunner: localRunner,
		semaphore:   semaphore.NewWeighted(int64(concurrency)),
	}
}

func (es *emailSenderTaskRunner) Run(ctx context.Context, wg *sync.WaitGroup) error {
	go func() {
		defer wg.Done()

		err := receiveMessage(ctx, es.conn, es.semaphore, sendEmailQueue, func(ctx context.Context, message *SendEmailMessage) error {
			err := es.localRunner.SendEmail(ctx, message.Address, message.Subject, message.Html, message.Text)
			if err != nil {
				return nil
			}

			return nil
		})
		if err != nil {
			slog.ErrorContext(ctx, "failed to receive message from queue", "error", err)
		}
	}()
	return nil
}
