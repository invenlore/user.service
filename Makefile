.PHONY: fresh

fresh:
	docker-compose down --remove-orphans -v
	docker-compose up -d --build

fresh/scaled:
	docker-compose down --remove-orphans -v
	docker-compose up -d --build --scale user-service=2

run:
	go build -o ./bin/user-service
	./bin/user-service

up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

workflow:
	pin-github-action .github/workflows/master.yaml
