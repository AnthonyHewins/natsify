# NATSify

Push on NATS, publish on NTFY. When a particular thing happens on my message bus that
I want to be notified about via my phone, the best OSS option is NTFY, rather than
have 100 NTFY clients floating around all my binaries this can sit in one place.
The main use case is when an error happens and pushing that error directly to my
phone. This micro-app just saves a bunch of duplication and it's a lot easier to
get instant notifications of something failing, or something happening that you
deem important

## Quickly run

To compile, `make natsify`. You can use flags or the equivalent `UPPER_SNAKE` of the below flags

```
Usage of natsify:
  -app-name="": Application name to include in all logs. Blank to exclude
  -log-exporter="": File to log to. Blank for stdout
  -log-format="json": Log format to use: json | logfmt
  -log-level="info": log level to use: debug | info | warn | error
  -log-src=false: Whether to include the line of source code that caused the log in the message
  -nats-max-reconnects=60: Override the max number of reconnect attempts. If negative, it will never stop trying to reconnect; defaults to 60
  -nats-timeout=2s: Override the default dial timeout on NATS
  -nats-topic="natsify": What topic to listen to
  -nats-url="nats://127.0.0.1:4222": Override the default NATS url
  -ntfy-timeout=5s: How long before sending a NTFY message will time out
  -ntfy-url="http://localhost:32016": Override the default NTFY url
```

I personally run this as a systemd service so it's always going. Edit the service file under the `systemd` directory
to your liking, then **if you are comfortable restarting the daemon**: `make deploy`

## Special case of error publishing

Since the main use case is error publishing, I use this a lot to push errors on NATS, so there is a client specifically
created for taking errors and pushing them on NATS