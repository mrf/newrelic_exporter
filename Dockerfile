FROM golang:1.19 AS builder

WORKDIR /app

COPY . .

RUN go mod download \
  && go get -v -d \
  && CGO_ENABLED=0 go build -o newrelic_exporter

FROM gcr.io/distroless/static-debian11

WORKDIR /app

COPY --from=builder /app/newrelic_exporter .

EXPOSE 9126

ENTRYPOINT ["/app/newrelic_exporter"]
