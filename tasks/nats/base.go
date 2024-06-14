package nats

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"golang.org/x/sync/semaphore"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const maxRetries = 5

type protobufPointer[B any] interface {
	protoreflect.ProtoMessage
	*B
}

func publishMessage(ctx context.Context, conn *nats.Conn, subject string, message protoreflect.ProtoMessage, retries int32) error {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	slog.Debug("Publishing message to NATS", "subject", subject, "retries", retries, "message", message, "messageType", fmt.Sprintf("%T", message))

	data, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	cdata, err := proto.Marshal(&MessageHeader{
		Metadata: carrier,
		Retries:  retries,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal MessageHeader: %w", err)
	}

	l := int32(len(cdata))
	finalMessage := append([]byte{byte(l >> 24), byte(l >> 16), byte(l >> 8), byte(l)}, cdata...)

	return conn.Publish(subject, append(finalMessage, data...))
}

func parseMessage[T any, PT protobufPointer[T]](ctx context.Context, msg *nats.Msg) (context.Context, *MessageHeader, *T, error) {
	l := int32(msg.Data[0])<<24 | int32(msg.Data[1])<<16 | int32(msg.Data[2])<<8 | int32(msg.Data[3])

	var header MessageHeader
	if err := proto.Unmarshal(msg.Data[1:1+l], &header); err != nil {
		return ctx, nil, nil, fmt.Errorf("failed to unmarshal MessageHeader: %w", err)
	}

	ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(header.Metadata))

	message := new(T)
	pt := PT(message)
	if err := proto.Unmarshal(msg.Data[1+l:], pt); err != nil {
		return ctx, nil, nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return ctx, &header, message, nil
}

func receiveMessage[T any, PT protobufPointer[T]](
	origCtx context.Context,
	conn *nats.Conn,
	semaphore *semaphore.Weighted,
	queue string,
	run func(context.Context, PT) error,
) error {
	subscription, err := conn.QueueSubscribeSync(runServerRemoteQueue, queue)
	if err != nil {
		return err
	}

	for {
		msg, err := subscription.NextMsgWithContext(origCtx)
		if err != nil {
			return err
		}

		slog.Debug("Received message from NATS", "subject", msg.Subject, "data", msg.Data)

		ctx, header, message, err := parseMessage[T, PT](origCtx, msg)
		if err != nil {
			return err
		}

		slog.Debug("Parsed message", "header", header, "message", message, "messageType", fmt.Sprintf("%T", message))

		if err := semaphore.Acquire(ctx, 1); err != nil {
			return err
		}

		done := make(chan struct{})
		go func() {
			defer semaphore.Release(1)
			defer close(done)

			msg.InProgress()

			slog.Debug("Running message", "message", message, "messageType", fmt.Sprintf("%T", message))

			pt := PT(message)
			err := run(ctx, pt)
			if err != nil {
				slog.Error("failed to run saver remote", "error", err)
				if header.Retries < maxRetries {
					msg, ok := interface{}(message).(protoreflect.ProtoMessage)
					if !ok {
						slog.Error("message is not a proto message")
						return
					}
					err := publishMessage(ctx, conn, queue, msg, header.Retries+1)
					if err != nil {
						slog.Error("failed to republish message", "error", err)
					}
				} else {
					slog.Info("message exceeded max retries", "retries", header.Retries)
				}
			}
		}()

		go func() {
			select {
			case <-ctx.Done():
				msg.Nak()
			case <-done:
				msg.Ack()
				return
			case <-time.Tick(5 * time.Second):
				msg.InProgress()
			}
		}()

		select {
		case <-ctx.Done():
			return errors.Join(ctx.Err(), subscription.Drain())
		default:
		}
	}
}
