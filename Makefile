ADDLICENSE=addlicense
ALL_SRC := $(shell find . -name '*.go' -o -name '*.sh' -type f | sort)

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

ifeq ($(BINDPLANE_VERSION), module)
BINDPLANE_VERSION := $(shell go list -m all | grep github.com/observiq/bindplane-op-enterprise | awk '{print $$2}')
endif

ifeq ($(GOOS), windows)
EXT?=.exe
else
EXT?=
endif

ifeq ($(GOARCH), amd64)
GOARCH_FULL?=amd64_v1
else
GOARCH_FULL=$(GOARCH)
endif

PWD=$(shell pwd)

# All source code and documents, used when checking for misspellings
ALLDOC := $(shell find . \( -name "*.md" -o -name "*.yaml" \) -type f | sort)

.PHONY: install-tools
install-tools:
	go install github.com/securego/gosec/v2/cmd/gosec@v2.16.0
	go install github.com/google/addlicense@v1.1.0
	go install github.com/mgechev/revive@v1.3.1
	go install github.com/uw-labs/lichen@v0.1.7
	go install github.com/goreleaser/goreleaser@v1.18.2
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/client9/misspell/cmd/misspell@v0.3.4

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: provider
provider:
	goreleaser build \
		--skip-validate \
		--single-target \
		--snapshot \
		--rm-dist \
		--config release/goreleaser.yml

.PHONY: release-test
release-test:
	goreleaser release \
		--skip-publish \
		--skip-validate \
		--snapshot \
		--rm-dist \
		--skip-sign \
		--config release/goreleaser.yml

.PHONY: ci-check
ci-check: lint misspell check-fmt gosec vet test check-license

.PHONY: lint
lint:
	revive -config .revive.toml -formatter friendly -set_exit_status ./...

.PHONY: misspell
misspell:
	misspell $(ALLDOC)

.PHONY: misspell-fix
misspell-fix:
	misspell -w $(ALLDOC)

.PHONY: check-fmt
check-fmt:
	goimports -d ./ | diff -u /dev/null -

.PHONY: fmt
fmt:
	goimports -w .

.PHONY: gosec
gosec:
	gosec -exclude-generated -exclude-dir internal/tools ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test:
	go test ./... -cover -race

.PHONY: test-cover
test-cover: vet
	go test ./... -race -cover -coverprofile cover.out
	go tool cover -html=cover.out

.PHONY: test-integration
test-integration: dev-tls
	@BINDPLANE_VERSION=$(BINDPLANE_VERSION) go test ./... --tags=integration -cover

.PHONY: test-integration-cover
test-integration-cover: dev-tls
	@BINDPLANE_VERSION=$(BINDPLANE_VERSION) go test ./... --tags=integration -cover -coverprofile cover.out
	go tool cover -html=cover.out

.PHONY: test-end-to-end
test-end-to-end: test-integration provider
	mkdir -p test/integration/providers
	cp dist/provider_$(GOOS)_$(GOARCH_FULL)/terraform-provider-bindplane* test/integration/providers/terraform-provider-bindplane_v0.0.0
	@BINDPLANE_VERSION=$(BINDPLANE_VERSION) bash test/integration/test.sh

# Test local configures test/local directory
# with the provider.
# Usage: After running this target, cd to test/local
# and run `export TF_CLI_CONFIG_FILE=./dev.tfrc` followed
# by your `terraform commands`.
.PHONY: test-local
test-local: provider
	rm -rf test/local/providers
	mkdir -p test/local/providers
	cp dist/$(GOOS)_$(GOARCH)/provider_$(GOOS)_$(GOARCH_FULL)/terraform-provider-bindplane* test/local/providers/terraform-provider-bindplane_v0.0.0


.PHONY: check-license
check-license:
	@ADDLICENSEOUT=`$(ADDLICENSE) -check $(ALL_SRC) 2>&1`; \
		if [ "$$ADDLICENSEOUT" ]; then \
			echo "$(ADDLICENSE) FAILED => add License errors:"; \
			echo "$$ADDLICENSEOUT"; \
			echo "Use 'make add-license' to fix this."; \
			exit 1; \
		else \
			echo "Check License finished successfully"; \
		fi

.PHONY: add-license
add-license:
	@ADDLICENSEOUT=`$(ADDLICENSE) -y "" -c "observIQ, Inc." $(ALL_SRC) 2>&1`; \
		if [ "$$ADDLICENSEOUT" ]; then \
			echo "$(ADDLICENSE) FAILED => add License errors:"; \
			echo "$$ADDLICENSEOUT"; \
			exit 1; \
		else \
			echo "Add License finished successfully"; \
		fi

dev-tls: client/tls
client/tls:
	mkdir client/tls
	docker run \
		-v ${PWD}/test/scripts/generate-dev-certificates.sh:/generate-dev-certificates.sh \
		-v ${PWD}/client/tls:/tls \
		--entrypoint=/bin/sh \
		alpine/openssl /generate-dev-certificates.sh

.PHONY: clean-dev-tls
clean-dev-tls:
	rm -rf client/tls
