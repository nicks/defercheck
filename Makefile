.PHONY: all check

check: lint test

test:
	go test ./internal/...

lint:
	go vet ./...
	errcheck ./...

install:
	go install github.com/nicks/defercheck
