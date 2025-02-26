# syntax=docker/dockerfile:1

FROM golang:1.24 as builder

ARG CGO_ENABLED=0
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o ./user-service

FROM scratch

COPY --from=builder /app/user-service /user-service
ENTRYPOINT ["/user-service"]
