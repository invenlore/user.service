run:
	go build -o ./bin/user-service cmd/main.go
	./bin/user-service

build:
	docker build --tag user-service .

compose:
	docker compose up --build -d

.PHONY: compose
