PROJECT ?= gitlab.similarweb.io/infrastructure/bbox
UNFORMATTED_FILES=$(shell find . -not -path "./vendor/*" -name "*.go" | xargs gofmt -s -l)
GOCMD=go
GOTEST=$(GOCMD) test
GOBUILD=$(GOCMD) build
GOARCH=amd64

.PHONY: release
release: ## Release locally all version
	@goreleaser release --clean --config .goreleaser.yml
