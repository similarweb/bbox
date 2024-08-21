.PHONY: test lint help version

NAME ?= bbox
GOCMD=go
GOTEST=$(GOCMD) test

test:  ## Run tests for the project
		$(GOTEST) ./... -v

lint: ## Run linter
	golangci-lint run 

help: ## Show Help menu
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

version: ## Show bbox current version
	$(GOCMD) run bbox version

