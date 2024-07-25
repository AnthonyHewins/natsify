package natsify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/AnthonyHewins/gotfy"
	"github.com/nats-io/nats.go"
)

var (
	ErrCantPushNil = errors.New("the message you tried to push was nil")
	ErrEmptySubj   = errors.New("empty subject passed to NATS")
	ErrEmptyTopic  = errors.New("empty topic passed to ntfy")
)

type Publisher struct {
	nc                            *nats.Conn
	appName, errSubject, errTopic string
}

func NewErrPublisher(ns *nats.Conn, appName, natsErrSubject, ntfyErrTopic string) (*Publisher, error) {
	if ns == nil || appName == "" || natsErrSubject == "" || ntfyErrTopic == "" {
		return nil, fmt.Errorf(
			"passed zero value to NewErrPublisher: ns: %v | app: %s | natsErrSubj: %s | ntfyErrTopic: %s",
			ns,
			appName,
			natsErrSubject,
			ntfyErrTopic,
		)
	}

	return &Publisher{
		nc:         ns,
		appName:    appName,
		errSubject: natsErrSubject,
		errTopic:   ntfyErrTopic,
	}, nil
}

func (e *Publisher) Push(ctx context.Context, subj string, m *gotfy.Message) error {
	switch {
	case subj == "":
		return ErrEmptySubj
	case m == nil:
		return ErrCantPushNil
	case m.Topic == "":
		return ErrEmptyTopic
	}

	buf, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return e.nc.Publish(subj, buf)
}

func (e *Publisher) PushErr(ctx context.Context, err error) error {
	if err == nil {
		return ErrCantPushNil
	}

	text := []string{err.Error()}
	for unwrapped := errors.Unwrap(err); unwrapped != nil; unwrapped = errors.Unwrap(unwrapped) {
		text = append(text, unwrapped.Error())
	}

	msg := gotfy.Message{
		Topic:   e.errTopic,
		Title:   e.appName + " error",
		Message: strings.Join(text, "\n"),
		Tags:    []string{gotfy.Red_circle},
	}

	buf, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return e.nc.Publish(e.errSubject, buf)
}
