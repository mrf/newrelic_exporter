FROM golang:1.16.15 AS builder

COPY . /app

RUN cd /app \
  && go mod init github.com/klinux/newrelic_exporter \
  && go get -v -d \
  && CGO_ENABLED=0 go build -o newrelic_exporter

FROM alpine:latest

RUN apk update && apk upgrade && apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/newrelic_exporter .

EXPOSE 9126

ENTRYPOINT ["/app/newrelic_exporter"]
