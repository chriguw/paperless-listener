IMAGE ?= paperless-listener
TAG ?= local
PLATFORMS ?= linux/amd64,linux/arm64
BUILDX_BUILDER ?= paperless-builder
MULTIARCH_OUTPUT ?= dist/multiarch-image.oci.tar
LIMA_SOCKET ?= $(HOME)/.lima/docker/sock/docker.sock

IMAGE_REF := $(IMAGE):$(TAG)

# Auto-detect Lima Docker socket if DOCKER_HOST is not already set
ifeq ($(DOCKER_HOST),)
  ifneq ($(wildcard $(LIMA_SOCKET)),)
    export DOCKER_HOST = unix://$(LIMA_SOCKET)
  endif
endif

.PHONY: help test lint clean build docker-build docker-run compose-up buildx-setup buildx-amd64 buildx-arm64 docker-buildx-local buildx-multiarch

help: ## Show available make targets
	@grep -E '^[a-zA-Z0-9_.-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "%-22s %s\n", $$1, $$2}'

test: ## Run Go tests
	go test ./...

lint: ## Run static checks
	go vet ./...

clean: ## Remove local build artifacts
	rm -rf bin dist

build: ## Build local binary to bin/paperless-listener
	mkdir -p bin
	go build -o bin/paperless-listener ./cmd/paperless-listener

docker-build: ## Build Docker image for current platform
	docker build -t $(IMAGE_REF) .

docker-run: ## Run Docker image with local config.json mounted read-only
	docker run --rm -p 8080:8080 -v "$(PWD)/config.json:/app/config.json:ro" $(IMAGE_REF)

compose-up: ## Start service via docker compose
	docker compose up --build

buildx-setup: ## Create/select buildx builder and bootstrap it
	docker buildx create --name $(BUILDX_BUILDER) --use 2>/dev/null || true
	docker buildx use $(BUILDX_BUILDER)
	docker buildx inspect --bootstrap

buildx-amd64: buildx-setup ## Build and load AMD64 image locally
	docker buildx build --platform linux/amd64 -t $(IMAGE):amd64 --load .

buildx-arm64: buildx-setup ## Build and load ARM64 image locally
	docker buildx build --platform linux/arm64 -t $(IMAGE):arm64 --load .

docker-buildx-local: buildx-setup ## Build multi-arch image locally to OCI archive (no push)
	mkdir -p $(dir $(MULTIARCH_OUTPUT))
	docker buildx build --platform $(PLATFORMS) -t $(IMAGE_REF) --output=type=oci,dest=$(MULTIARCH_OUTPUT) .

buildx-multiarch: buildx-setup ## Build and push multi-arch image
	docker buildx build --platform $(PLATFORMS) -t $(IMAGE_REF) --push .

