PROJECT_NAME := "aci-chatbot"
PKG_LIST := $(shell go list ./... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ )

fmt: ## Format the files
	.ci/fmt.sh;

test: ## Run unittests
	@go test -v -short ${PKG_LIST}

coverage: ## Generate global code coverage report
	.ci/coverage.sh;

coverhtml: ## Generate global code coverage report in HTML
	.ci/coverage.sh html;

dep: ## Get the dependencies
	@go get -v -d ./...

build: dep ## Build the binary file
	@go build -o $(PROJECT_NAME)
