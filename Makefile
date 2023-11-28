# Go Command Path
GO ?= `which go`

# GoRelease Command Path
GORELEASER ?= `which goreleaser`

GREEN = "\\033[0\;32m"
BOLD = "\\033[1m"
RESET = "\\033[0m"

.PHONY: help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "make \033[36m%-10s\033[0m %s\n", $$1, $$2}'

.PHONY: generate
generate: ## Generates the openhue.gen.go client from the latest https://github.com/openhue/openhue-api specification
	@$(GO) generate
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
test: ## Tidy makes sure go.mod matches the source code in the module
	@$(GO) test ./... -coverprofile=c.out
	@echo "\n${GREEN}${BOLD}all tests successfully passed ✅ ${RESET}"


.PHONY: clean
clean: ##
	@rm -rf ./dist
	@$(GO) clean
	@echo "\n${GREEN}${BOLD}Project successfully cleaned 🧹${RESET} (removed ./dist folder + go clean)"