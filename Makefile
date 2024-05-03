LOCAL_BIN:=$(CURDIR)/bin
GOLANGCI_TAG:=1.56.0

# install golangci-lint binary
.PHONY: install-lint
install-lint:
ifeq ($(wildcard $(GOLANGCI_BIN)),)
	$(info Downloading golangci-lint v$(GOLANGCI_TAG))
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_TAG)
GOLANGCI_BIN:=$(LOCAL_BIN)/golangci-lint
endif

.PHONY:
tests:
	go test -coverprofile=coverage.out ./...

.PHONY:
show_cover:
	go tool cover -html=coverage.out

.PHONY:
linter:
	$(LOCAL_BIN)/golangci-lint run --fix

.PHONY:
quality: linter tests

.PHONY:
clean_cache:
	go clean -cache

.PHONY:
build:
	go build -o easyp ./cmd
