package natsify

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/AnthonyHewins/gotfy"
	"github.com/nats-io/nats.go"
)

var (
	ErrCantPushNil = errors.New("the error you tried to push was nil")
)

type ErrPublisher struct {
	nc             *nats.Conn
	natsSubject    string
	appName, topic string
}

func NewErrPublisher(ns *nats.Conn, natsSubject, ntfyTopic, appName string) *ErrPublisher {
	return &ErrPublisher{
		nc:      ns,
		appName: appName,
		topic:   ntfyTopic,
	}
}

func (e *ErrPublisher) Push(ctx context.Context, err error) error {
	if err == nil {
		return ErrCantPushNil
	}

	text := []string{err.Error()}
	for unwrapped := errors.Unwrap(err); unwrapped != nil; unwrapped = errors.Unwrap(unwrapped) {
		text = append(text, unwrapped.Error())
	}

	msg := gotfy.Message{
		Topic:   e.topic,
		Title:   e.appName + " error",
		Message: strings.Join(text, "\n"),
		Tags:    []string{gotfy.Red_circle},
	}

	buf, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return e.nc.Publish(e.natsSubject, buf)
}
