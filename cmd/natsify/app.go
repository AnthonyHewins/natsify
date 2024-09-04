package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AnthonyHewins/gotfy"
	"github.com/nats-io/nats.go"
)

func (a *app) read(ctx context.Context, msg *nats.Msg) (*gotfy.Message, error) {
	l := a.logger.With("subject", msg.Subject)

	l.DebugContext(ctx, "received message")

	var x gotfy.Message
	if err := json.Unmarshal(msg.Data, &x); err != nil {
		l.ErrorContext(ctx, "failed reading message", "err", err, "bytes", string(msg.Data))
		return nil, err
	}

	l = a.logger.With("msg", x)

	if x.Topic == "" {
		l.ErrorContext(ctx, "invalid topic passed: empty string")
		return nil, fmt.Errorf("no topic received: empty string")
	}

	return &x, nil
}

func (a *app) publish(ctx context.Context, m *gotfy.Message) error {
	resp, err := a.publisher.SendMessage(ctx, m)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed publishing message", "err", err, "resp", resp)
		return err
	}

	a.logger.DebugContext(ctx, "published message", "resp", resp)
	return nil
}
