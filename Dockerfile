# syntax=docker/dockerfile:1

FROM golang:1.24 AS builder

ARG CGO_ENABLED=0
WORKDIR /app

COPY . ./

RUN go mod download
RUN go build -o ./bin/service

FROM scratch

WORKDIR /app

COPY --from=builder /app/bin/service ./service
ENTRYPOINT ["/app/service"]
