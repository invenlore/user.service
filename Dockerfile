# syntax=docker/dockerfile:1

FROM golang:1.24 AS builder

ARG CGO_ENABLED=0
WORKDIR /app

RUN go env -w GOMODCACHE=/root/.cache/go-build

COPY . ./

RUN --mount=type=cache,target=/root/.cache/go-build go mod download
RUN --mount=type=cache,target=/root/.cache/go-build go build -o ./bin/service

FROM scratch

WORKDIR /app

COPY --from=builder /app/bin/service ./service
ENTRYPOINT ["/app/service"]
