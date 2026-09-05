# GHP developer commands — the local equivalent of the CI pipeline
# (see .github/workflows/ci.yml). Every check CI runs is available here.

GO          := go
BIN         := bin
COVERAGE    := coverage.out
COVERAGE_MIN:= 90

.PHONY: help setup fmt lint vet \
	test test-short test-full test-integration test-e2e \
	coverage coverage-html \
	build build-all \
	vscode-test \
	gate clean

help: ## List all targets with a short description
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'

setup: ## Enable the git hooks (gofmt+vet on commit, Conventional Commit check on commit-msg)
	git config core.hooksPath .githooks

fmt: ## Reformat all Go source with gofmt
	gofmt -w ./src

lint: ## Formatting check (gofmt -l, no output = ok) + go vet
	@unformatted=$$(gofmt -l ./src); \
	if [ -n "$$unformatted" ]; then \
		echo "Formatting violations:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi
	$(GO) vet ./src/...

vet: ## Run go vet on the whole Go tree
	$(GO) vet ./src/...

test-short: ## Unit tests only (skips integration/e2e via -short)
	$(GO) test -short ./src/... -race

test: test-short ## Alias for test-short

test-integration: ## Integration tests only
	$(GO) test ./src/test/integration/... -race

test-e2e: build ## E2E tests against the compiled bin/ghp binary
	GHP_BINARY=$(CURDIR)/$(BIN)/ghp $(GO) test ./src/test/e2e/...

test-full: ## Everything: unit + integration + e2e
	$(GO) test ./src/... -race

coverage: ## Full test run with coverage + enforce the 90% minimum
	$(GO) test ./src/... -coverprofile=$(COVERAGE) -covermode=atomic -coverpkg=./src/...
	@total=$$($(GO) tool cover -func=$(COVERAGE) | grep total: | awk '{print $$3}'); \
	pct=$${total%%%}; \
	echo "Total coverage: $$total"; \
	awk -v c=$$pct -v m=$(COVERAGE_MIN) 'BEGIN {exit !(c>=m)}' || { \
		echo "ERROR: coverage $$pct% is below the required $(COVERAGE_MIN)%;"; \
		exit 1; \
	}

coverage-html: coverage ## Open the per-line coverage report in a browser
	$(GO) tool cover -html=$(COVERAGE)

build: ## Build the ghp binary into bin/
	mkdir -p $(BIN)
	$(GO) build -o $(BIN)/ghp ./src/cmd/ghp

build-all: ## Compile every package (same as the CI Build job)
	$(GO) build ./src/...

vscode-test: ## Run the VS Code extension tests (editors/vscode, npm)
	cd editors/vscode && npm test

gate: lint test-full coverage build-all ## Everything the CI gate checks

clean: ## Remove build artifacts and coverage output
	rm -rf $(BIN) $(COVERAGE)