ifneq ("$(wildcard .env)","")
  $(info using .env)
  include .env
  export $(shell sed 's/=.*//' .env)
endif

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: run/api
run/api:
	@go run ./cmd...

current_time = $(shell date "+%Y-%m-%dT%H:%M:%S%z")
git_description = $(shell git describe --always --dirty --tags --long)
linker_flags = '-s -X main.buildTime=${current_time} -X main.version=${git_description}'

ISALPINE := $(shell grep 'Alpine' /etc/os-release  -c)
musl=
ifeq ($(ISALPINE), 2)
        musl=-tags musl
endif

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd...'	
	go build ${musl} -ldflags=${linker_flags} -o=./bin/api ./cmd

## clean/apps: clear generated bin files
.PHONY: clean/apps
clean/apps:
	@echo 'Remove builded apps'
	@rm -rf ./bin

## docker/build: build the local environment for development
.PHONY: docker/build
docker/build:
	docker-compose up --build

## docker/up: start the local stack in background
.PHONY: docker/up
docker/up:
	docker-compose up -d

## docker/down: shutdown the running containers
.PHONY: docker/down
docker/down:
	docker-compose down	

## test/local: local test all code
.PHONY: local/test
test/local:
	go test -race -vet=off -coverpkg ./... -v -coverprofile=cover.out ./...
	go tool cover -html=cover.out 

## test: test all code
.PHONY: test
test:
	go test -race -vet=off -coverpkg ./... -v -coverprofile=cover.out ./...
	go tool cover -func=cover.out

## generate: generate mocks
.PHONY: generate
generate:
	ROOT_DIR=$(shell pwd) go generate ./...

## audit: tidy dependencies, format and vet all code
.PHONY: audit
audit:
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	golangci-lint run

## tidy: tidy dependencies
.PHONY: tidy
tidy:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	golangci-lint run --fix