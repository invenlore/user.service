.PHONY: fresh

fresh:
	docker-compose down --remove-orphans -v
	docker-compose up -d --build

fresh/scaled:
	docker-compose down --remove-orphans -v
	docker-compose up -d --build --scale identity-service=2

run:
	go build -o ./bin/identity-service
	./bin/identity-service

restart:
	docker-compose down --remove-orphans
	docker-compose up -d --build

up:
	docker-compose up -d --build

down:
	docker-compose down --remove-orphans

logs:
	docker-compose logs -f

workflow:
	pin-github-action .github/workflows/master.yaml
