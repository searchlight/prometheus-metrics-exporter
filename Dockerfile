# stage 1: build
FROM golang:1.10-alpine AS builder
LABEL maintainer="nightfury1204"

# Add source code
RUN mkdir -p /go/src/github.com/searchlight/prometheus-metrics-exporter
ADD . /go/src/github.com/searchlight/prometheus-metrics-exporter

# Build binary
RUN cd /go/src/github.com/searchlight/prometheus-metrics-exporter && \
    GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/prometheus-remote-metric-writer

# stage 2: lightweight "release"
FROM alpine:latest
LABEL maintainer="nightfury1204"

COPY --from=builder /go/bin/prometheus-remote-metric-writer /bin/

ENTRYPOINT [ "/bin/prometheus-remote-metric-writer" ]
