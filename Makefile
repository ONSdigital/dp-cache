.PHONY: all
all: audit test build

.PHONY: audit
audit:
	go list -json -m all | nancy sleuth

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
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.43.0
	golangci-lint run ./...

.PHONY: test
test:
	go test -race -cover ./...