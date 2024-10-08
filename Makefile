GOLANGCI_LINT_VERSION=v1.60.1
GOLANG_VULCHECK_VERSION=v1.1.3

# TODO(@JOJO) im not sure if this is the best way to organize the variables related to azure functions
functions-folder=functions
esaj-api-function=esaj-api
# To connect with the firestore emulator, you need to set the FIRESTORE_EMULATOR_HOST environment variable. Using the same host and port defined in the inicialization of the container
# More: https://pkg.go.dev/cloud.google.com/go/firestore#hdr-Google_Cloud_Firestore_Emulator
export FIRESTORE_EMULATOR_HOST=0.0.0.0:8087

## build the esaj service for local usage. Usage `make esaj`
.PHONY: esaj
esaj:
	go build -o ./cmd/esaj/ ./cmd/esaj

## build the esaj function to run in the azure OS. Usage `make esaj-function`
.PHONY: esaj-function
esaj-function:
	GOOS=linux GOARCH=amd64 go build -o ./functions ./cmd/esaj

## Run the pdf to markdown service from cmd/pdf2markdown/main.go
.PHONY: pdf2markdown
pdf2markdown:
	go run cmd/pdf2markdown/main.go

## run all tests. Usage `make test` or `make test testcase="TestFunctionName"` to run an isolated tests
.PHONY: test
test:
	if [ -n "$(testcase)" ]; then \
		go test ./... -timeout 10s -race -run="^$(testcase)$$" -v; \
	else \
		go test ./... -timeout 10s -race; \
	fi

## Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./... -v
	go run golang.org/x/vuln/cmd/govulncheck@$(GOLANG_VULCHECK_VERSION) ./...

## Run test coverage
.PHONY: coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

## Run integration tests. Usage `make integration-test` or `make integration-test testcase="TestFunctionName"` to run an isolated tests
.PHONY: integration-test
integration-test:
	@echo "Running integration tests..."
	if [ -n "$(testcase)" ]; then \
		go test ./... -tags integration -timeout 10s -v -run="^$(testcase)$$" ; \
	else \
		go test ./... -tags integration -timeout 10s; \
	fi

## Start the development server
.PHONY: dev/start
dev/start:
	@echo "Starting the development server..."
	@docker-compose up -d

 ## Stop the development server
.PHONY: dev/stop
dev/stop:
	@echo "Stopping the development server..."
	@docker-compose down

## Generate the mocks for the tests
.PHONY: generate-mocks
generate-mocks:
	@echo "Generating mocks..."
	go generate ./...

## Display help for all targets
.PHONY: help
help:
	@awk '/^.PHONY: / { \
		msg = match(lastLine, /^## /); \
			if (msg) { \
				cmd = substr($$0, 9, 100); \
				msg = substr(lastLine, 4, 1000); \
				printf "  ${GREEN}%-30s${RESET} %s\n", cmd, msg; \
			} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)
