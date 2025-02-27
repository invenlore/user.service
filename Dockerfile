# syntax=docker/dockerfile:1

FROM golang:1.24 AS builder

ARG CGO_ENABLED=0
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go build -o ./user-service cmd/main.go

FROM scratch

COPY --from=builder /app/user-service /user-service
ENTRYPOINT ["/user-service"]
