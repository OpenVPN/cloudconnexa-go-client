.PHONE: all
all: test

.PHONY: deps
deps:
	@go mod download

.PHONY: test e2e lint build clean

test:
	go test -v -race ./cloudconnexa/...

e2e:
	go test -v -race ./e2e/...

lint:
	golangci-lint run

build:
	go build -v ./...

clean:
	go clean
	rm -f coverage.txt
