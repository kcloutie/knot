IS_WINDOWS=false
OUTPUT_FILE=knot
BUILD_DATE=NA
BUILD_COMMIT=$(shell git rev-parse HEAD)
BUILD_VERSION=$(shell git describe --abbrev=0 --tags)
BUILD_DATE="$$(date --iso=seconds)"
SERVER_CONFIG_FILE=test/files/serverConfig.json
GO_TEST_FLAGS +=

TIMEOUT_UNIT = 20m
TIMEOUT_E2E  = 20m

ifeq ($(OS),Windows_NT)
# SHELL := powershell.exe
# .SHELLFLAGS := -NoProfile -Command
IS_WINDOWS=true
OUTPUT_FILE=knot.exe
else



endif
LDFLAGS += -ldflags "-X github.com/kcloutie/knot/pkg/params/version.BuildTime=$(BUILD_DATE) -X github.com/kcloutie/knot/pkg/params/version.BuildVersion=$(BUILD_VERSION) -X github.com/kcloutie/knot/pkg/params/version.Commit=$(BUILD_COMMIT)"

.PHONY: build
build:
	@echo ""
	@echo "Is Windows: ${IS_WINDOWS}"
	@echo "BUILD_DATE: ${BUILD_DATE}"
	@echo "BUILD_COMMIT: ${BUILD_COMMIT}"
	@echo "BUILD_VERSION: ${BUILD_VERSION}"
	@echo "OUTPUT_FILE: ${OUTPUT_FILE}"
	@echo ""
	@echo "Building CLI..."
	@go build $(LDFLAGS) -o bin/$(OUTPUT_FILE) cmd/knot/knot.go

.PHONY: api-server
api-server: build
	@echo "Running Server..."
	@./$(OUTPUT_FILE) run server -c $(SERVER_CONFIG_FILE)

TEST_UNIT_TARGETS := test-unit-verbose test-unit-race test-unit-failfast
test-unit-verbose: ARGS=-v
test-unit-failfast: ARGS=-failfast
test-unit-race:    ARGS=-race
$(TEST_UNIT_TARGETS): test-unit
test-clean:  ## Clean testcache
	@echo "Cleaning test cache"
	@go clean -testcache 
.PHONY: $(TEST_UNIT_TARGETS) test test-unit
test: test-clean test-unit ## Run test-unit
test-unit: ## Run unit tests
	@echo "Running unit tests..."
	@go test $(GO_TEST_FLAGS) -timeout $(TIMEOUT_UNIT) $(ARGS) -cover ./... 

.PHONY: test-e2e-cleanup
test-e2e-cleanup: ## cleanup test e2e namespace/pr left open
	@echo "Cleaning e2e tests..."
	# @./hack/dev/Stop-TestServer.ps1

.PHONY: test-e2e
test-e2e:  test-e2e-cleanup ## run e2e tests
	@go test $(GO_TEST_FLAGS) -timeout $(TIMEOUT_E2E)  -failfast -count=1 -tags=e2e $(GO_TEST_FLAGS) ./test

.PHONY: release
release: unit-test
	@echo "Releasing Product..."
	@echo "TAG: ${TAG}"
	@git tag -a $(Tag) -m "Release ${TAG}"
	@goreleaser --rm-dist

.PHONY: docs
docs:
	@echo "Generating Docs..."
	@go run ./cmd/gen-docs --standard --doc-path docs/knot

.PHONY: docs-custom
docs-custom:
	@echo "Generating Custom Docs..."
	@go run ./cmd/gen-docs --custom --doc-path docs/knot-custom