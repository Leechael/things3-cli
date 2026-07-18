BINARY_NAME := things3-cli
BIN_DIR := bin
BIN_PATH := $(BIN_DIR)/$(BINARY_NAME)
OUT_DIR := dist
CMD := ./cmd/things3-cli
GOFLAGS ?= -buildvcs=false

# Version embedding configuration
# Standard location: pkg/version.Version or similar
# Set to the full package path of the Version variable
# We extract it from release-naming.env later if possible
VERSION_VAR := $(shell grep VERSION_VAR_PATH release-naming.env 2>/dev/null | cut -d= -f2)
ifeq ($(VERSION_VAR),)
	VERSION_VAR := github.com/Leechael/things3--cli/pkg/version.Version
endif
GIT_COMMIT_VAR := $(shell grep GIT_COMMIT_VAR_PATH release-naming.env 2>/dev/null | cut -d= -f2)
ifeq ($(GIT_COMMIT_VAR),)
	GIT_COMMIT_VAR := github.com/Leechael/things3--cli/pkg/version.GitCommit
endif
BUILD_TIME_VAR := $(shell grep BUILD_TIME_VAR_PATH release-naming.env 2>/dev/null | cut -d= -f2)
ifeq ($(BUILD_TIME_VAR),)
	BUILD_TIME_VAR := github.com/Leechael/things3--cli/pkg/version.BuildTime
endif
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X $(VERSION_VAR)=$(VERSION) -X $(GIT_COMMIT_VAR)=$(GIT_COMMIT) -X $(BUILD_TIME_VAR)=$(BUILD_TIME)

export GOFLAGS

.PHONY: tidy fmt test bdd-test ci build run install clean cross-build help

help:
	@echo "Targets:"
	@echo "  make tidy         - go mod tidy"
	@echo "  make fmt          - gofmt ./cmd ./internal ./tests"
	@echo "  make test         - run unit tests"
	@echo "  make bdd-test     - run BDD tests"
	@echo "  make ci           - fmt check + vet + tests + build"
	@echo "  make build        - build local binary to ./$(BIN_PATH)"
	@echo "  make run          - run CLI (pass ARGS='...')"
	@echo "  make install      - install binary to GOPATH/bin"
	@echo "  make clean        - remove build artifacts"
	@echo "  make cross-build  - build darwin/linux amd64/arm64 binaries"

tidy:
	go mod tidy

fmt:
	gofmt -w ./cmd ./internal ./tests

test:
	go test ./... -count=1

bdd-test:
	go test -tags=bdd ./tests/bdd/... -count=1

ci:
	@unformatted=$$(gofmt -l ./cmd ./internal ./tests); \
	if [ -n "$$unformatted" ]; then echo "Unformatted files:"; echo "$$unformatted"; exit 1; fi
	go vet ./...
	go test ./... -count=1
	go test -tags=bdd ./tests/bdd/... -count=1
	go build -v -ldflags="$(LDFLAGS)" -o $(BIN_PATH) $(CMD)

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_PATH) -ldflags="$(LDFLAGS)" $(CMD)

run:
	go run -ldflags="$(LDFLAGS)" $(CMD) $(ARGS)

install:
	go install -ldflags="$(LDFLAGS)" $(CMD)

clean:
	rm -rf $(OUT_DIR) $(BIN_DIR)

cross-build: clean
	mkdir -p $(OUT_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(OUT_DIR)/$(BINARY_NAME)-darwin-amd64 -ldflags="$(LDFLAGS)" $(CMD)
	GOOS=darwin GOARCH=arm64 go build -o $(OUT_DIR)/$(BINARY_NAME)-darwin-arm64 -ldflags="$(LDFLAGS)" $(CMD)
	GOOS=linux GOARCH=amd64 go build -o $(OUT_DIR)/$(BINARY_NAME)-linux-amd64 -ldflags="$(LDFLAGS)" $(CMD)
	GOOS=linux GOARCH=arm64 go build -o $(OUT_DIR)/$(BINARY_NAME)-linux-arm64 -ldflags="$(LDFLAGS)" $(CMD)
