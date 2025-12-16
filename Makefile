.PHONY: fresh

fresh:
	docker-compose down --remove-orphans
	docker-compose up -d --build -V

run:
	go build -o ./bin/user-service
	./bin/user-service

up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f
