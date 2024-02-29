PROJECT ?= gitlab.similarweb.io/infrastructure/bbox
UNFORMATTED_FILES=$(shell find . -not -path "./vendor/*" -name "*.go" | xargs gofmt -s -l)
GOCMD=go
GOTEST=$(GOCMD) test
GOBUILD=$(GOCMD) build
GOARCH=amd64

.PHONY: release
release: ## Release locally all version
	@goreleaser release --clean --config .goreleaser.yml

test:  ## Run tests for the project
		$(GOTEST) -short -race -cover -failfast ./...

fmt: ## Format Go code
	@gofmt -w -s main.go $(UNFORMATTED_FILES)

lint:
	golangci-lint run  --timeout=3m

fmt-check: ## Check go code formatting
	@echo "==> Checking that code complies with gofmt requirements..."
	@if [ ! -z "$(UNFORMATTED_FILES)" ]; then \
		echo "gofmt needs to be run on the following files:"; \
		echo "$(UNFORMATTED_FILES)" | xargs -n1; \
		echo "You can use the command: \`make fmt\` to reformat code."; \
		exit 1; \
	else \
		echo "Check passed."; \
	fi

help: ## Show Help menu
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

