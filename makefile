NAME ?= bbox
PROJECT ?= https://github.com/similarweb/bbox
GOCMD=go
GOTEST=$(GOCMD) test

test:  ## Run tests for the project
		$(GOTEST)  ./... -v

lint: ## Run linter for the project
	golangci-lint run 

help: ## Show Help menu
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


.PHONY: version
version: ## Print the version of the project
	$(GOCMD) run $(NAME) version 

.PHONY: version describe
version-describe: ## Print the version of the project with full description
	$(GOCMD) run $(NAME) version -d 
