.PHONY: all test build

all: test build

test: ## Run unittests
	@go test

build: ## Build the binary file
	@go build -i -o usr/bin/bizfly-agent *.go

clean: ## Remove previous build
	@rm -rf usr/bin/bizfly-agent
