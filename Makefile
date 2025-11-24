.PHONY: all
all: audit lint test build

.PHONY: audit
audit:
	dis-vulncheck

.PHONY: build
build:
	go build ./...

.PHONY: convey
convey:
	goconvey ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.0
	golangci-lint run ./...

.PHONY: test
test:
	go test -race -cover ./...
