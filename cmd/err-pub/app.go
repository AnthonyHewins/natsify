package main

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/AnthonyHewins/gotfy"
	"github.com/nats-io/nats.go"
)

type app struct {
	ns        *nats.Subscription
	publisher *gotfy.Publisher
	logger    *slog.Logger
}

func (a *app) read(ctx context.Context) (*gotfy.Message, error) {
	msg, err := a.ns.NextMsgWithContext(ctx)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed reading message", "err", err)
		return nil, err
	}

	var x gotfy.Message
	if err := json.Unmarshal(msg.Data, &x); err != nil {
		a.logger.ErrorContext(ctx, "failed reading message", "err", err)
		return nil, err
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
