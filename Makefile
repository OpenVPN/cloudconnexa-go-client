.PHONE: all
all: test

.PHONY: deps
deps:
	@go mod download

.PHONY: test
test:
	@go test -cover -race -v ./...

e2e:
	@go test -v ./e2e/...