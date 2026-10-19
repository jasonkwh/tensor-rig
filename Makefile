.PHONY: lint test-unit

lint:
	golangci-lint run -c .golangci.yml

test-unit:
	go test -count=1 ./...
