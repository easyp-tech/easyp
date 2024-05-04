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

.PHONY: bin-deps
bin-deps:
	$(info Installing binary dependencies...)
	mkdir -p $(LOCAL_BIN)
	GOBIN=$(LOCAL_BIN) go mod tidy && \
	GOBIN=$(LOCAL_BIN) go install github.com/vektra/mockery/v2@v2.41.0

.PHONY: install
install: bin-deps install-lint

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
	go build -o easyp ./cmd/easyp

.PHONY:
mockery:
	$(LOCAL_BIN)/mockery --name $(name) --dir $(dir) --output $(dir)/mocks

.PHONY:
mock:
	make mockery name=LockFile dir=./internal/mod/adapters/storage
