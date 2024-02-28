FROM golang:1.22 AS builder

ARG VERSION=dev

COPY . /go/src/app
WORKDIR /go/src/app

RUN CGO_ENABLED=0 GOOS=linux go build -o connector -ldflags="-X 'main.version=$VERSION'" main.go

FROM alpine:3.19

RUN mkdir -p /opt/connector
WORKDIR /opt/connector
COPY --from=builder /go/src/app/connector connector

ENTRYPOINT ["./connector"]
