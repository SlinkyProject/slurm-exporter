##@ General

# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= 0.2.1

# Image URL to use all building/pushing image targets
IMG ?= slinky.slurm.net/slurm-exporter:$(VERSION)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Get the OS to set platform specific commands
UNAME_S ?= $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	CP_FLAGS = -v -n
else
	CP_FLAGS = -v --update=none
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build

.PHONY: all
all: build ## Build slurm-exporter.

REGISTRY ?= slinky.slurm.net

.PHONY: build
build: ## Build manager binary.
	$(MAKE) docker-build IMG="$(REGISTRY)/slurm-exporter:$(VERSION)"
	helm package helm/slurm-exporter

.PHONY: push
push: build ## Push container images.
	$(MAKE) docker-push IMG="$(REGISTRY)/slurm-exporter:$(VERSION)"
	$(foreach chart, $(wildcard ./*.tgz), helm push ${chart} oci://$(REGISTRY)/charts ;)

.PHONY: clean
clean: ## Clean executable files.
	@ chmod -R -f u+w bin/ || true # make test installs files without write permissions.
	rm -rf bin/
	rm -f govulnreport.txt
	rm -f cover.out cover.html
	rm -f *.tgz

.PHONY: run
run: fmt tidy vet ## Run the exporter from your host.
	go run ./cmd/main.go

# If you wish built the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64 ). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	$(CONTAINER_TOOL) build $(CONTAINER_BUILD_FLAGS) -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push ${IMG}

# PLATFORMS defines the target platforms for  the manager image be build to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - able to use docker buildx . More info: https://docs.docker.com/build/buildx/
# - have enable BuildKit, More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image for your registry (i.e. if you do not inform a valid value via IMG=<myregistry/image:<tag>> then the export will fail)
# To properly provided solutions that supports more than one platform you should use this option.
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: test ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- $(CONTAINER_TOOL) buildx create --name project-v3-builder
	$(CONTAINER_TOOL) buildx use project-v3-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile.cross .
	- $(CONTAINER_TOOL) buildx rm project-v3-builder
	rm Dockerfile.cross

##@ Deployment

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

##@ Development

.PHONY: install-dev
install-dev: ## Install binaries for development environment.
	go install github.com/norwoodj/helm-docs/cmd/helm-docs@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/go-delve/delve/cmd/dlv@latest
	go install sigs.k8s.io/kind@latest

.PHONY: helm-dependency-update
helm-dependency-update: ## Update Helm chart dependencies.
	find "helm/" -depth -mindepth 1 -maxdepth 1 -type d -print0 | xargs -0r -n1 helm dependency update

.PHONY: values-dev
values-dev: ## Safely initialize values-dev.yaml files for Helm charts.
	find "helm/" -type f -name "values.yaml" | sed 'p;s/\.yaml/-dev\.yaml/' | xargs -n2 cp $(CP_FLAGS)

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: tidy
tidy: ## Run go mod tidy against code
	go mod tidy

.PHONY: get-u
get-u: ## Run `go get -u`
	go get -u ./...
	$(MAKE) tidy

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

CODECOV_PERCENT ?= 80

.PHONY: test
test: ## Run tests.
	rm -f cover.out cover.html
	go test -v -coverprofile cover.out ./...
	go tool cover -func cover.out
	go tool cover -html cover.out -o cover.html
	@percentage=$$(go tool cover -func=cover.out | grep ^total | awk '{print $$3}' | tr -d '%'); \
		if (( $$(echo "$$percentage < $(CODECOV_PERCENT)" | bc -l) )); then \
			echo "----------"; \
			echo "Total test coverage ($${percentage}%) is less than the coverage threshold ($(CODECOV_PERCENT)%)."; \
			exit 1; \
		fi

.PHONY: vuln-scan
vuln-scan: ## Run vulnerability scanning tool
	GOBIN=$(LOCALBIN) go install golang.org/x/vuln/cmd/govulncheck@latest
	$(LOCALBIN)/govulncheck ./... > govulnreport.txt 2>&1 || echo "Found vulnerabilities. Details in govulnreport.txt"

.PHONY: audit
audit: fmt tidy vet vuln-scan
