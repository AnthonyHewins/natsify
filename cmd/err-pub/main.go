package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/namsral/flag"

	"github.com/AnthonyHewins/gotfy"
	"github.com/nats-io/nats.go"
)

type config struct {
	ntfyURL     string
	ntfyTimeout time.Duration

	natsTopic         string
	natsURL           string
	natsTimeout       time.Duration
	natsMaxReconnects int

	logAppName  string
	logExporter string
	logLevel    string
	logFmt      string
	logSrc      bool
}

func newApp() (*app, error) {
	var c config

	flag.StringVar(&c.ntfyURL, "ntfy-url", "http://localhost:32016", "Override the default NTFY url")
	flag.DurationVar(&c.ntfyTimeout, "ntfy-timeout", time.Second*5, "How long before sending a NTFY message will time out")

	flag.StringVar(&c.natsTopic, "nats-topic", "", "What topic to listen to; required")
	flag.StringVar(&c.natsURL, "nats-url", "nats://127.0.0.1:4222", "Override the default NATS url")
	flag.DurationVar(&c.natsTimeout, "nats-timeout", time.Second*2, "Override the default dial timeout on NATS")
	flag.IntVar(&c.natsMaxReconnects, "nats-max-reconnects", 60, "Override the max number of reconnect attempts. If negative, it will never stop trying to reconnect; defaults to 60")

	flag.StringVar(&c.logAppName, "app-name", "", "Application name to include in all logs. Blank to exclude")
	flag.StringVar(&c.logExporter, "log-exporter", "", "File to log to. Blank for stdout")
	flag.StringVar(&c.logLevel, "log-level", "info", "log level to use: debug | info | warn | error")
	flag.StringVar(&c.logFmt, "log-format", "json", "Log format to use: json | logfmt")
	flag.BoolVar(&c.logSrc, "log-src", false, "Whether to include the line of source code that caused the log in the message")

	flag.Parse()

	logger, err := c.slogger()
	if err != nil {
		return nil, err
	}

	publisher, err := c.ntfy(logger)
	if err != nil {
		return nil, err
	}

	ns, err := c.natsConn(logger)
	if err != nil {
		return nil, err
	}

	return &app{
		ns:        ns,
		publisher: publisher,
		logger:    logger,
	}, nil
}

func main() {
	a, err := newApp()
	if err != nil {
		panic(err)
	}

	for {
		msg, err := a.read(context.Background())
		if err != nil {
			panic(err)
		}

		if err = a.publish(context.Background(), msg); err != nil {
			panic(err)
		}
	}
}

func (c *config) ntfy(logger *slog.Logger) (*gotfy.Publisher, error) {
	l := logger.With("url", c.ntfyURL, "timeout", c.ntfyTimeout)

	l.Debug("creating NTFY client...")
	ntfyURL, err := url.Parse(c.ntfyURL)
	if err != nil {
		logger.Error(fmt.Sprintf("failed parsing NTFY URL: %s", *ntfyURL), "err", err)
		return nil, err
	}

	publisher, err := gotfy.NewPublisher(ntfyURL, &http.Client{Timeout: c.ntfyTimeout})
	if err != nil {
		logger.Error(
			"failed connecting to NTFY; outputting config",
			"url", ntfyURL,
			"timeout", c.ntfyTimeout,
		)
		return nil, err
	}
	logger.Debug("client created (cannot test connection until first message)")

	return publisher, nil
}

func (c *config) natsConn(logger *slog.Logger) (*nats.Subscription, error) {
	l := logger.With(
		"url", c.natsURL,
		"max reconnects", c.natsMaxReconnects,
		"timeout", c.natsTimeout,
		"topic", c.natsTopic,
	)

	l.Info("connecting to NATS...")
	nc, err := nats.Connect(
		c.natsURL,
		nats.MaxReconnects(c.natsMaxReconnects),
		nats.Timeout(c.natsTimeout),
	)

	if err != nil {
		logger.Error(
			"failed connecting to nats; outputting config",
		)
		return nil, err
	}
	logger.Info("connected to NATS successfully, subbing to topic...")

	ns, err := nc.SubscribeSync(c.natsTopic)
	if err != nil {
		logger.Error("failed subscribing to the topic specified", "err", err)
		return nil, err
	}

	logger.Info("subbed to topic")
	return ns, nil
}

type Reader struct {
	Nc        *nats.Subscription
	Publisher *gotfy.Publisher
}

func (r *Reader) NextMsg(ctx context.Context) error {
	msg, err := r.Nc.NextMsgWithContext(ctx)
	if err != nil {
		return err
	}

	var x gotfy.Message
	if err = json.Unmarshal(msg.Data, &x); err != nil {
		return err
	}

	_, err = r.Publisher.SendMessage(ctx, &x)
	return err
}

func (c *config) slogger() (*slog.Logger, error) {
	logLevel, exporter, logFmt, addSrc, appName := c.logLevel, c.logExporter, c.logFmt, c.logSrc, c.logAppName

	var level slog.HandlerOptions
	switch strings.ToLower(logLevel) {
	case "":
		return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.Level(math.MaxInt)})), nil
	case "debug":
		level = slog.HandlerOptions{Level: slog.LevelDebug, AddSource: addSrc}
	case "info":
		level = slog.HandlerOptions{Level: slog.LevelInfo, AddSource: addSrc}
	case "warn":
		level = slog.HandlerOptions{Level: slog.LevelWarn, AddSource: addSrc}
	case "err":
		level = slog.HandlerOptions{Level: slog.LevelError, AddSource: addSrc}
	default:
		return nil, fmt.Errorf("invalid log level: %s", logLevel)
	}

	out, err := getExporter(exporter)
	if err != nil {
		return nil, err
	}

	var logger *slog.Logger
	switch logFmt {
	case "", "json":
		logger = slog.New(slog.NewJSONHandler(out, &level))
	case "text", "logfmt":
		logger = slog.New(slog.NewTextHandler(out, &level))
	default:
		return nil, fmt.Errorf("invalid handler format: %s", logFmt)
	}

	if appName != "" {
		logger = logger.With("app-name", appName)
	}

	return logger, nil
}

func getExporter(exporter string) (io.Writer, error) {
	switch exporter {
	case "":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	}

	file, err := os.Create(exporter)
	if err != nil {
		return nil, err
	}

	return file, nil
}
