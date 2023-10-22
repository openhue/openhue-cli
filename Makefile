# Go Command Path
GO ?= `which go`

GREEN = "\\033[0\;32m"
BOLD = "\\033[1m"
RESET = "\\033[0m"

.PHONY: help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "make \033[36m%-10s\033[0m %s\n", $$1, $$2}'

.PHONY: generate
generate: ## Generates the openhue.gen.go client from the latest https://github.com/openhue/openhue-api specification
	@$(GO) generate
	@echo "${GREEN}${BOLD}./openhue/openhue.gen.go successfully generated 🚀${RESET}"

.PHONY: build
build: ## Generates the openhue-cli executable, and output it in the ./bin folder
	@$(GO) build -o bin/openhue
	@echo "${GREEN}${BOLD}openhue binary successfully generated in ./bin folder 📦${RESET}"
	@echo "Checking if the binary file is valid...\n"
	./bin/openhue -h

.PHONY: tidy
tidy: ## Tidy makes sure go.mod matches the source code in the module
	@$(GO) mod tidy
	@echo "${GREEN}${BOLD}go.mod successfully cleaned 🧽${RESET}"

.PHONY: clean
clean: ##
	@rm -rf ./bin
	@$(GO) clean
	@echo "${GREEN}${BOLD}Project successfully cleaned 🧹${RESET} (removed ./bin folder + go clean)"