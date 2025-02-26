build:
	docker build --tag user-service .

compose:
	docker compose up --build

.PHONY: compose
