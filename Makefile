run:
	go build -o ./bin/user-service
	./bin/user-service

up:
	docker compose up -d

down:
	docker compose down

fresh:
	docker compose down --remove-orphans
	docker compose build --no-cache
	docker compose up -d --build -V

logs:
	docker compose logs -f

.PHONY: fresh
