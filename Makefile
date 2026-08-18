.PHONY: default up lint test-unit

default: up

up:
	pulumi up

lint:
	golangci-lint run -c .golangci.yml

test-unit:
	go test -count=1 ./...
