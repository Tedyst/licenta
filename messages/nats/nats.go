package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/messages"
)

type NATSExchange struct {
	conn *nats.Conn
}

func NewNATSExchange(conn *nats.Conn) (*NATSExchange, error) {
	return &NATSExchange{conn: conn}, nil
}

func (n *NATSExchange) PublishSendScanToWorkerMessage(ctx context.Context, worker *queries.Worker, message messages.SendScanToWorkerMessage) error {
	subject := fmt.Sprintf("worker.%d", worker.ID)
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return n.conn.Publish(subject, msgBytes)
}

func (n *NATSExchange) ReceiveSendScanToWorkerMessage(ctx context.Context, worker *queries.Worker) (messages.SendScanToWorkerMessage, bool, error) {
	subject := fmt.Sprintf("worker.%d", worker.ID)
	sub, err := n.conn.SubscribeSync(subject)
	if err != nil {
		return messages.SendScanToWorkerMessage{}, false, err
	}
	defer sub.Unsubscribe()

	msg, err := sub.NextMsgWithContext(ctx)
	if err != nil {
		if err == nats.ErrTimeout {
			return messages.SendScanToWorkerMessage{}, false, nil
		}
		return messages.SendScanToWorkerMessage{}, false, err
	}

	var message messages.SendScanToWorkerMessage
	if err := json.Unmarshal(msg.Data, &message); err != nil {
		return messages.SendScanToWorkerMessage{}, false, err
	}

	return message, true, nil
}
