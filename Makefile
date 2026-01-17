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

up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

workflow:
	pin-github-action .github/workflows/master.yaml
