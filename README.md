# bizfly-agent

[![Go Report Card](https://goreportcard.com/badge/github.com/VCCloud/bizfly-agent)](https://goreportcard.com/report/github.com/VCCloud/bizfly-agent)

Collect system metrics and send to pushgateway.

## Building

With go module enable `export GO111MODULE=on`:

```sh
$ go build
$ ./bizfly-agent
```

## Note

`bizfly-agent` uses node exporter, with some modification to filesystem metrics to report the whole volume instead of mount points.

Only **Linux** is supported.
