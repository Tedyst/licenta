package local

import (
	"context"
	"reflect"
	"strconv"
	"sync"

	"github.com/tedyst/licenta/messages"
	"github.com/tedyst/licenta/models"
)

const bufferSize = 100

type localExchange struct {
	channels map[string]chan interface{}
	mutex    sync.Mutex
}

func NewLocalExchange() *localExchange {
	return &localExchange{
		channels: make(map[string]chan interface{}),
	}
}

func (e *localExchange) getWorkerTopic(worker models.Worker) string {
	return "worker." + strconv.Itoa(int(worker.ID))
}

func (e *localExchange) PublishSendScanToWorkerMessage(ctx context.Context, worker *models.Worker, message *messages.SendScanToWorkerMessage) error {
	return e.publish(ctx, e.getWorkerTopic(*worker), message)
}

func (e *localExchange) ReceiveSendScanToWorkerMessage(ctx context.Context, worker *models.Worker) (messages.SendScanToWorkerMessage, bool, error) {
	message := messages.SendScanToWorkerMessage{}
	ok, err := e.receive(ctx, e.getWorkerTopic(*worker), &message)
	return message, ok, err
}

func (e *localExchange) publish(ctx context.Context, topic string, message interface{}) error {
	e.mutex.Lock()

	channel, ok := e.channels[topic]
	if !ok {
		channel = make(chan interface{}, bufferSize)
		e.channels[topic] = channel
	}

	e.mutex.Unlock()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case channel <- message:
		return nil
	}
}

func (e *localExchange) receive(ctx context.Context, topic string, message interface{}) (bool, error) {
	channel, ok := e.channels[topic]
	if !ok {
		channel = make(chan interface{}, bufferSize)
		e.channels[topic] = channel
	}
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case msg := <-channel:
		reflect.ValueOf(message).Elem().Set(reflect.ValueOf(msg))
		return true, nil
	}
}
