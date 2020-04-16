# Makefile

GO_PREFIX ?= github.com/virtual-vgo/vvgo

# Quick build cmds
.PHONY: vvgo vvgo-uploader # Use go build tools caching
BIN_PATH ?= .
BUILD_FLAGS ?= -v
vvgo:
	go generate ./... && go build -v -o $(BIN_PATH)/$@ $(GO_PREFIX)/cmd/vvgo
vvgo-uploader:
	go generate ./... && go build -v -o $(BIN_PATH)/$@ $(GO_PREFIX)/cmd/vvgo-uploader

# Generate code
generate: cmd/vvgo/info.go cmd/vvgo-uploader/info.go data/statik/statik.go
cmd/vvgo/info.go:
	go generate $(GO_PREFIX)/cmd/vvgo
cmd/vvgo-uploader/info.go:
	go generate $(GO_PREFIX)/cmd/vvgo-uploader
data/statik/statik.go: data
	go generate $(GO_PREFIX)/data

# Run tests
.PHONY: fmt vet test
fmt:
	gofmt -d .
vet:
	go vet $(GO_PREFIX)/...
TEST_FLAGS ?= -race
test: generate fmt vet
	go test $(TEST_FLAGS) $(GO_PREFIX)/...

# Make releases

HARDWARE ?= $(shell uname -m)
RELEASE_TAG ?= $(shell git rev-parse --short HEAD)

.PHONY: releases releases/$(BIN_PATH) releases/$(IMAGE_REPO)
releases: releases/$(BIN_PATH)
releases/$(BIN_PATH): $(BIN_PATH)/vvgo-uploader-$(RELEASE_TAG)-linux-$(HARDWARE)
releases/$(BIN_PATH): $(BIN_PATH)/vvgo-uploader-$(RELEASE_TAG)-darwin-$(HARDWARE)
releases/$(BIN_PATH): $(BIN_PATH)/vvgo-uploader-$(RELEASE_TAG)-windows-$(HARDWARE).exe

$(BIN_PATH)/%-$(RELEASE_TAG)-darwin-$(HARDWARE):
	GOOS=darwin go build -v -o $(BIN_PATH)/$@ $(GO_PREFIX)/cmd/vvgo-uploader
$(BIN_PATH)/%-$(RELEASE_TAG)-linux-$(HARDWARE):
	GOOS=linux go build -v -o $(BIN_PATH)/$@ $(GO_PREFIX)/cmd/vvgo-uploader
$(BIN_PATH)/%-$(RELEASE_TAG)-windows-$(HARDWARE).exe:
	GOOS=windows go build -v -o $(BIN_PATH)/$@ $(GO_PREFIX)/cmd/vvgo-uploader

# Build images

VVGO_IMAGE_NAME=vvgo
PAGE_CACHE_IMAGE_NAME=page-cache
OBJECT_CACHE_IMAGE_NAME=object-cache
KV_CACHE_IMAGE_NAME=kv-cache

GITHUB_REPO ?= virtual-vgo/vvgo
IMAGE_REPO ?= docker.pkg.github.com/$(GITHUB_REPO)
VVGO_IMAGE_CACHE ?= $(IMAGE_REPO)/$(VVGO_IMAGE_NAME):latest
PAGE_CACHE_IMAGE_CACHE ?= $(IMAGE_REPO)/$(PAGE_CACHE_IMAGE_NAME):latest
OBJECT_CACHE_IMAGE_CACHE ?= $(IMAGE_REPO)/$(OBJECT_CACHE_IMAGE_NAME):latest
KV_CACHE_IMAGE_CACHE ?= $(IMAGE_REPO)/$(KV_CACHE_IMAGE_NAME):latest

.PHONY: images
images: images/page-cache
images: images/object-cache
images: images/kv-cache
images: images/vvgo-builder
images: images/vvgo

images/vvgo-builder:
	docker pull $(VVGO_IMAGE_CACHE)-builder || true
	docker build . \
		--file Dockerfile \
		--target builder \
		--cache-from=$(VVGO_IMAGE_CACHE)-builder \
		--build-arg GITHUB_SHA=$GITHUB_SHA \
		--build-arg GITHUB_REF=$GITHUB_REF \
		--tag vvgo-builder:$(RELEASE_TAG)

images/vvgo:
	docker pull $(VVGO_IMAGE_CACHE) || true
	docker build . \
		--file Dockerfile \
		--target vvgo \
		--cache-from=$(VVGO_IMAGE_CACHE)-builder \
		--cache-from=$(VVGO_IMAGE_CACHE) \
		--tag vvgo:$(RELEASE_TAG)

images/page-cache:
	docker pull $(PAGE_CACHE_IMAGE_CACHE) || true
	docker build infra/object-cache \
		--file infra/object-cache/Dockerfile \
		--cache-from=$(PAGE_CACHE_IMAGE_CACHE) \
		--tag page-cache:$(RELEASE_TAG)

images/object-cache:
	docker pull $(OBJECT_CACHE_IMAGE_CACHE) || true
	docker build infra/page-cache \
		--file infra/page-cache/Dockerfile \
		--cache-from=$(OBJECT_CACHE_IMAGE_CACHE) \
		--tag object-cache:$(RELEASE_TAG)

images/kv-cache:
	docker pull $(KV_CACHE_IMAGE_CACHE) || true
	docker build infra/kv-cache \
		--file infra/kv-cache/Dockerfile \
		--cache-from=$(KV_CACHE_IMAGE_CACHE) \
		--tag kv-cache:$(RELEASE_TAG)

# Deploy images

.PHONY: deploy push/$(IMAGE_REPO)
deploy: push/$(IMAGE_REPO)

deploy/$(IMAGE_REPO): push/$(IMAGE_REPO)/page-cache\:$(RELEASE_TAG)
deploy/$(IMAGE_REPO): push/$(IMAGE_REPO)/object-cache\:$(RELEASE_TAG)
deploy/$(IMAGE_REPO): push/$(IMAGE_REPO)/kv-cache\:$(RELEASE_TAG)
deploy/$(IMAGE_REPO): push/$(IMAGE_REPO)/vvgo-builder\:$(RELEASE_TAG)
deploy/$(IMAGE_REPO): push/$(IMAGE_REPO)/vvgo\:$(RELEASE_TAG)

deploy/$(IMAGE_REPO)/%\:$(RELEASE_TAG): images/%
	docker tag $*:$(RELEASE_TAG) $(IMAGE_REPO)/$*:$(RELEASE_TAG)
	docker push $(IMAGE_REPO)/$*
