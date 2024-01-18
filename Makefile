# Go Command Path
GO ?= `which go`

# GoRelease Command Path
GORELEASER ?= `which goreleaser`

GREEN = "\\033[0\;32m"
BOLD = "\\033[1m"
RESET = "\\033[0m"

SPEC_URL = "https://api.redocly.com/registry/bundle/openhue/openhue/v2/openapi.yaml?branch=main"

.PHONY: help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "make \033[36m%-10s\033[0m %s\n", $$1, $$2}'

.PHONY: generate
generate: ## Generates the openhue/gen/openhue.gen.go client. Usage: make generate [spec=/path/to/openhue.yaml]
ifdef spec
	@echo "Code generation from $(spec)"
	@oapi-codegen --package=gen -generate=client,types -o ./openhue/gen/openhue.gen.go "$(spec)"
else
	@echo "Code generation from $(SPEC_URL)"
	@oapi-codegen --package=gen -generate=client,types -o ./openhue/gen/openhue.gen.go "$(SPEC_URL)"
endif
	@echo "\n${GREEN}${BOLD}./openhue/openhue.gen.go successfully generated 🚀${RESET}"

.PHONY: build
build: ## Generates the openhue-cli executables in the ./dist folder
	@$(GORELEASER) check
	@$(GORELEASER) build --clean --snapshot --single-target
	@echo "\n${GREEN}${BOLD}openhue binaries successfully generated in the ./dist folder 📦${RESET}"

.PHONY: tidy
tidy: ## Tidy makes sure go.mod matches the source code in the module
	@$(GO) mod tidy
	@echo "\n${GREEN}${BOLD}go.mod successfully cleaned 🧽${RESET}"

.PHONY: test
test: ## Run the tests
	@$(GO) test ./...
	@echo "\n${GREEN}${BOLD}all tests successfully passed ✅ ${RESET}"

.PHONY: coverage
coverage: ## Run the tests with coverage. Usage: make coverage [html=true]
	@$(GO) test ./... -coverprofile=c.out
ifdef html
	@$(GO) tool cover -html="c.out"
else
	@$(GO) tool cover -func="c.out"
endif
	@echo "\n${GREEN}${BOLD}all tests successfully passed ✅ ${RESET}"

.PHONY: clean
clean: ##
	@rm -rf ./dist
	@rm -f c.out
	@rm -f coverage.html
	@$(GO) clean
	@echo "\n${GREEN}${BOLD}Project successfully cleaned 🧹${RESET} (removed ./dist folder + go clean)"