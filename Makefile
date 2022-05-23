PROJECT_NAME ?= robot

ROOT_DIR=$(PWD)

TINKOFF_PROTO=$(ROOT_DIR)/investapi


all:
	@echo "build			- Build project"
	@echo "setup			- Setup project"
	@echo "setup-dev		- Setup dev dependencies"
	@echo "clean			- Remove compiled proto"
	@echo "compile-proto	- Compile all .proto files"
	@echo "lint				- Run linter"
	@echo "tests			- Run tests"
	@echo "coverage			- Show test's coverage"
	@exit 0

build:
	go build -v ./cmd/run-robot/
	go build -v ./cmd/generate-config/
	go build -v ./cmd/strategy-backtest/

setup:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go mod verify

setup-dev:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

clean:
	rm -f $(TINKOFF_PROTO)/*.go
	rm -f ./run-robot
	rm -f ./generate-config
	rm -f ./strategy-backtest

compile-proto: clean
	protoc -I=$(TINKOFF_PROTO) --go_out=$(TINKOFF_PROTO)/ --go-grpc_out=$(TINKOFF_PROTO)/ $(TINKOFF_PROTO)/*

lint:
	go vet ./...
	golangci-lint run

tests:
	go test -race -v -failfast ./...

coverage:
	go test -cover ./...

.PHONY: all build setup clean compile-proto lint tests coverage
