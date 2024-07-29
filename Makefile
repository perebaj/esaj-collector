GOLANGCI_LINT_VERSION = v1.59.1

## Run the esaj service from cmd/esaj/main.go
.PHONY: esaj
esaj:
	go run cmd/esaj/main.go

## Run the headless service from cmd/headless/main.go
.PHONY: headless
headless:
	go run cmd/headless/main.go

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
lint: ## Run linter
	@echo "Running linter..."
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./... -v

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
