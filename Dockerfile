FROM golang:1.18.3 AS builder

COPY . /app

RUN cd /app \
  && go get -v -d \
  && CGO_ENABLED=0 go build -o newrelic_exporter

FROM alpine:latest

RUN apk update && apk upgrade && apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/newrelic_exporter .

EXPOSE 9126

ENTRYPOINT ["/app/newrelic_exporter"]
