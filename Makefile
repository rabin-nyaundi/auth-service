## Load env variables from the .envrc file
include .envrc

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n "s/^##//p" ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# .PHONY: confirm
# confirm:
# 	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]

## run/api: runs the application
.PHONY: run/api
run/api:
	go run ./cmd/api -db-dsn=postgresql://db_admin:admin_21@localhost/user_db

## db/psql: Connect to the database using psql
.PHONY: db/sql
db/psql: 
	psql postgresql://db_admin:admin_21@localhost/user_db

## db/migrations/new name=$1: create a new databse migration
.PHONY: db/migrations/new
db/migrations/new:
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up migrations
.PHONY: db/migrations/up
db/migrations/up: ## confirm
	@echo 'Running up migrations'
	migrate -path ./migrations -database ${DATABASE_DSN} up

## db/migrations/up: apply all up migrations
.PHONY: db/migrations/down
db/migrations/down: confirm
	@echo 'Running down migrations'
	migrate -path ./migrations -database ${DATABASE_DSN} down

## audit: tidy dependencis and format, vet, and test all code
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	@go fmt ./...
	@echo 'Vetting code ...'
	@go vet ./...
	@staticcheck ./...
	@echo "running tests..."
	@go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	@go mod tidy
	@go mod verify
	@echo 'Vendoring dependencies'
	go mod vendor



current_time = $(shell date --iso-8601=seconds)
git_description = $(shell git describe --always --dirty --tags --long)
linkerFlags = '-s -X main.buildTime=${current_time} -X main.version=${git_description}'

## build/api: build the application
.PHONY: build/api
build/api:
	@echo "building cmd/api"
	go build -ldflags=${linkerFlags} -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags=${linkerFlags} -o=./bin/linux_amd64 ./cmd/api