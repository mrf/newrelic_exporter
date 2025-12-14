FROM --platform=$BUILDPLATFORM golang:1.25 AS builder

ARG TARGETOS
ARG TARGETARCH

COPY . /app

RUN cd /app \
  && go mod download \
  && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o newrelic_exporter

FROM alpine:3.20

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/newrelic_exporter .

EXPOSE 9126

ENTRYPOINT ["/app/newrelic_exporter"]
