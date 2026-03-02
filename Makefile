LOCALBIN       ?= $(shell pwd)/bin
GO             ?= go
GOLANGCI_LINT  ?= $(LOCALBIN)/golangci-lint
LINT_VERSION   ?= v2.10.1
GOFLAGS        := -race
TESTFLAGS      := -v -count=1

.PHONY: test lint lint-fix tidy verify-tidy build cover

test:
	$(GO) test $(GOFLAGS) $(TESTFLAGS) ./...

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run ./...

lint-fix: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --fix ./...

tidy:
	$(GO) mod tidy

verify-tidy:
	$(GO) mod tidy
	git diff --exit-code go.mod go.sum

build:
	$(GO) build ./...

cover:
	$(GO) test $(GOFLAGS) -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

$(GOLANGCI_LINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(LOCALBIN) $(LINT_VERSION)
