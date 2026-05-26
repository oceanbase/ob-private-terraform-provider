BINARY_NAME  = terraform-provider-oceanbase
GOBIN       ?= $(shell go env GOPATH)/bin
HOSTNAME     = registry.terraform.io
NAMESPACE    = oceanbase
TYPE         = oceanbase
VERSION     ?= 0.1.0

OS_ARCH     := $(shell go env GOOS)_$(shell go env GOARCH)
INSTALL_DIR  = $(HOME)/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(TYPE)/$(VERSION)/$(OS_ARCH)

.PHONY: build install uninstall test testacc lint clean fmt help

build: ## Build the provider binary
	go build -o $(BINARY_NAME) .

install: build ## Build and install to the local Terraform plugin directory
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY_NAME) $(INSTALL_DIR)/

uninstall: ## Remove the locally installed provider
	rm -rf $(HOME)/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(TYPE)

test: ## Run unit tests
	go test ./... -v

testacc: ## Run acceptance tests (requires OCP_URL / OCP_USERNAME / OCP_PASSWORD)
	TF_ACC=1 go test ./internal/provider/... -v -timeout 60m

lint: ## Run static analysis
	go vet ./...

fmt: ## Format source code
	go fmt ./...

clean: ## Remove build artifacts
	rm -f $(BINARY_NAME)
	rm -rf dist/

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
