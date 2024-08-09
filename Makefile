.PHONE: all
all: test

.PHONY: deps
deps:
	@go mod download

.PHONY: test
test:
	@go test -cover -race -v ./cloudconnexa/...

.PHONY: e2e
e2e:
	@go test -v ./e2e/...