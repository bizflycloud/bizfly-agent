PROJECT_NAME := "bizfly-agent"
PKG := "github.com/bizflycloud/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)


.PHONY: all test test-coverage build

all: test build

test: ## Run unittests
	@go test

test-coverage: ## Run tests with coverage
	@go test -short -coverprofile cover.out -covermode=atomic ${PKG_LIST}
	@cat cover.out >> coverage.txt

build: ## Build the binary file
	@go build -i -o usr/bin/bizfly-agent *.go

clean: ## Remove previous build
	@rm -rf usr/bin/bizfly-agent


help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
