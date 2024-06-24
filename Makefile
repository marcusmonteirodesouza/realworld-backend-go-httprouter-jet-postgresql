.PHONY: build/api
build/api:
	CGO_ENABLED=0 go build -ldflags '-s -w' -o ./bin/api ./cmd/api

.PHONY: confirm
confirm:
	@echo 'Are you sure? [y/N]' && read ans && [ $${ans:-N} = y ]

## db/migrations/down: password=$1 apply all down database migrations
.PHONY: db/migrations/down
db/migrations/down: confirm
	migrate -path="./migrations" -database "postgres://realworld:${password}@localhost/realworld?sslmode=disable" down

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: password=$1 apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	migrate -path="./migrations" -database "postgres://realworld:${password}@localhost/realworld?sslmode=disable" up

.PHONY: down
down:
	docker compose down

.PHONY: run/api
run/api:
	go run ./cmd/api

.PHONY: up
up:
	docker compose up --detach --wait

.PHONY: up/db
up/db:
	docker compose up postgres --detach --wait