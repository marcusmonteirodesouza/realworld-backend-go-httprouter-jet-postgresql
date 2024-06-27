.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags '-s -w' -o ./bin/api ./cmd/api

.PHONY: confirm
confirm:
	@echo 'Are you sure? [y/N]' && read ans && [ $${ans:-N} = y ]

## db/clear password=$1: clear and rebuild database and run Jet generator
.PHONY: db/clear
db/clear: confirm
	make down && rm -rf postgres-data && rm -rf .gen && make up/db/build && make jet password=${password}

## db/migrations/down password=$1: apply all down database migrations
.PHONY: db/migrations/down
db/migrations/down: confirm
	migrate -path="./migrations" -database "postgres://postgres:${password}@localhost/realworld?sslmode=disable" down

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up password=$1: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	migrate -path="./migrations" -database "postgres://postgres:${password}@localhost/realworld?sslmode=disable" up

## db/migrations/force password=$1 version=$2: force migration version
.PHONY: db/migrations/force
db/migrations/force: confirm
	migrate -path="./migrations" -database "postgres://postgres:${password}@localhost/realworld?sslmode=disable" force "${version}"

.PHONY: down
down:
	docker compose down

.PHONY: format
format:
	go fmt ./... && npx prettier --write .

## jet password=$1: run Jet generator 
.PHONY: jet
jet:
	jet -dsn="postgres://postgres:${password}@localhost/realworld?sslmode=disable" -schema=public -path=./.gen

.PHONY: run
run: up/db
	air

.PHONY: up
up:
	docker compose up --wait

.PHONY: up/build
up/build:
	docker compose up --wait --build

.PHONY: up/db
up/db:
	docker compose up migrations --exit-code-from migrations

.PHONY: up/db/build
up/db/build:
	docker compose up migrations --build --exit-code-from migrations