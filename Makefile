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

TOOLS_MOD_DIR := ./internal/tools

.PHONY: install-tools
install-tools:
	cd $(TOOLS_MOD_DIR) && go install github.com/securego/gosec/v2/cmd/gosec
	cd $(TOOLS_MOD_DIR) && go install github.com/google/addlicense
	cd $(TOOLS_MOD_DIR) && go install github.com/mgechev/revive
	cd $(TOOLS_MOD_DIR) && go install github.com/uw-labs/lichen
	cd $(TOOLS_MOD_DIR) && go install github.com/goreleaser/goreleaser/v2
	cd $(TOOLS_MOD_DIR) && go install github.com/client9/misspell/cmd/misspell

.PHONY: tidy
tidy:
	go mod tidy

# rm the dist/ directory because the `--rm-dist` flag will not
# if the dist/ directory already exists with directories and files
# that do not conflict with the output from the build command.
.PHONY: provider
provider:
	rm -rf dist/
	goreleaser build \
		--skip=validate \
		--single-target \
		--snapshot \
		--config release/goreleaser.yml

.PHONY: release-test
release-test:
	goreleaser release \
		--skip=publish \
		--skip=validate \
		--snapshot \
		--clean \
		--skip=sign \
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
	find dist -type f -name 'terraform-provider-bindplane*' -exec cp {} test/integration/providers/terraform-provider-bindplane_v0.0.0 \;
	@BINDPLANE_VERSION=$(BINDPLANE_VERSION) BINDPLANE_LICENSE=$(BINDPLANE_LICENSE) bash test/integration/test.sh

# Test local configures test/local directory
# with the provider.
# Usage: After running this target, cd to test/local
# and run `export TF_CLI_CONFIG_FILE=./dev.tfrc` followed
# by your `terraform commands`.
.PHONY: test-local
test-local: provider
	rm -rf test/local/providers
	mkdir -p test/local/providers
	find dist -type f -name 'terraform-provider-bindplane*' -exec cp {} test/local/providers/terraform-provider-bindplane_v0.0.0 \;

.PHONY: test-local-apply
test-local-apply:
	make test-local && cd test/local && terraform destroy -auto-approve && terraform apply -auto-approve && cd ../../


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

