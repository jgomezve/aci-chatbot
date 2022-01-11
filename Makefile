PROJECT_NAME := "aci-chatbot"
PKG_LIST := $(shell go list ./... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ )

fmt: ## Format the files
	@gofmt -d ${GO_FILES}

test: ## Run unittests
	@go test -short ${PKG_LIST}

coverage: ## Generate global code coverage report
	.ci/coverage.sh;

coverhtml: ## Generate global code coverage report in HTML
	.ci/coverage.sh html;

dep: ## Get the dependencies
	@go get -v -d ./...
	@go get -u github.com/golang/lint/golint

build: dep ## Build the binary file
	@go build -i -v $(PKG)
