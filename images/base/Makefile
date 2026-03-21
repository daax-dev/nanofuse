.PHONY: help build test clean push validate shell inspect all

# Image configuration
IMAGE_NAME ?= nanofuse-base
IMAGE_TAG ?= latest
REGISTRY ?= ghcr.io/jpoley
FULL_IMAGE_NAME = $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

# Build artifacts
BUILD_DIR = ./build
ROOTFS_FILE = $(BUILD_DIR)/rootfs.ext4
KERNEL_FILE = $(BUILD_DIR)/vmlinux
MANIFEST_FILE = $(BUILD_DIR)/manifest.json

help:
	@echo "NanoFuse Base Image Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build      - Build Docker image and extract rootfs + kernel (calls build.sh)"
	@echo "  validate   - Validate build artifacts (rootfs, kernel, manifest)"
	@echo "  test       - Test image boots in Firecracker (requires firecracker binary)"
	@echo "  clean      - Remove build artifacts and Docker images"
	@echo "  push       - Push Docker image to registry (GHCR)"
	@echo "  all        - Build, validate, and test"
	@echo "  shell      - Start interactive shell in Docker image"
	@echo "  inspect    - Show detailed information about build artifacts"
	@echo ""
	@echo "Configuration:"
	@echo "  IMAGE_NAME:  $(IMAGE_NAME)"
	@echo "  IMAGE_TAG:   $(IMAGE_TAG)"
	@echo "  REGISTRY:    $(REGISTRY)"
	@echo "  FULL_IMAGE:  $(FULL_IMAGE_NAME)"
	@echo ""
	@echo "Note: Build requires sudo for mounting ext4 filesystems"

all: build validate test

# Build calls build.sh (which requires sudo)
build:
	@if [ ! -x build.sh ]; then chmod +x build.sh; fi
	@IMAGE_NAME=$(IMAGE_NAME) IMAGE_TAG=$(IMAGE_TAG) ./build.sh

# Validate build artifacts
validate:
	@if [ ! -f "$(ROOTFS_FILE)" ] || [ ! -f "$(KERNEL_FILE)" ] || [ ! -f "$(MANIFEST_FILE)" ]; then \
		echo "ERROR: Build artifacts not found. Run 'make build' first."; \
		exit 1; \
	fi
	@echo "Validating build artifacts..."
	@chmod +x validate-build.sh
	@./validate-build.sh $(BUILD_DIR)

# Test boot in Firecracker
test:
	@if [ ! -f "$(ROOTFS_FILE)" ] || [ ! -f "$(KERNEL_FILE)" ]; then \
		echo "ERROR: Build artifacts not found. Run 'make build' first."; \
		exit 1; \
	fi
	@echo "Testing image boot in Firecracker..."
	@chmod +x test-boot.sh
	@if [ ! -x "$$(which firecracker 2>/dev/null)" ]; then \
		echo "ERROR: firecracker binary not found in PATH"; \
		echo "Install Firecracker: https://github.com/firecracker-microvm/firecracker/releases"; \
		exit 1; \
	fi
	@./test-boot.sh $(KERNEL_FILE) $(ROOTFS_FILE)

# Clean all build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@docker rm nanofuse-build-temp 2>/dev/null || true
	@docker rmi $(IMAGE_NAME):$(IMAGE_TAG) 2>/dev/null || true
	@echo "✓ Clean complete"

# Push Docker image to registry
push:
	@if [ ! -f "$(ROOTFS_FILE)" ]; then \
		echo "ERROR: Build artifacts not found. Run 'make build' first."; \
		exit 1; \
	fi
	@echo "Pushing image to registry..."
	@if [ -z "$(REGISTRY)" ]; then \
		echo "ERROR: REGISTRY not set"; \
		exit 1; \
	fi
	@docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(FULL_IMAGE_NAME)
	@docker push $(FULL_IMAGE_NAME)
	@echo "✓ Image pushed: $(FULL_IMAGE_NAME)"

# Start interactive shell in Docker image
shell:
	@echo "Starting interactive shell in image..."
	@docker run --rm -it $(IMAGE_NAME):$(IMAGE_TAG) /bin/bash

# Show detailed information about build artifacts
inspect:
	@if [ ! -f "$(ROOTFS_FILE)" ] || [ ! -f "$(KERNEL_FILE)" ] || [ ! -f "$(MANIFEST_FILE)" ]; then \
		echo "ERROR: Build artifacts not found. Run 'make build' first."; \
		exit 1; \
	fi
	@echo "=== Image Inspection ==="
	@echo ""
	@echo "Rootfs details:"
	@ls -lh $(ROOTFS_FILE)
	@file $(ROOTFS_FILE)
	@echo ""
	@echo "Kernel details:"
	@ls -lh $(KERNEL_FILE)
	@file $(KERNEL_FILE)
	@echo ""
	@echo "Manifest contents:"
	@cat $(MANIFEST_FILE) | jq . 2>/dev/null || cat $(MANIFEST_FILE)
