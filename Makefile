# Makefile

GO_PREFIX ?= github.com/virtual-vgo/vvgo

# Tests

.PHONY: fmt vet test
fmt:
	gofmt -d .

vet:
	go vet $(GO_PREFIX)/...

TEST_FLAGS ?= -race
test: fmt vet
	go test $(TEST_FLAGS) $(GO_PREFIX)/...

# Build vvgo

# Use go build tools caching, so mark this as a phony target
.PHONY: vvgo vvgo-uploader
BIN_PATH ?= .
BUILD_FLAGS ?= -v
vvgo:
	go generate ./... && go build -v -o $(BIN_PATH)/vvgo $(GO_PREFIX)/cmd/vvgo

vvgo-uploader:
	go generate ./... && go build -v -o $(BIN_PATH)/vvgo-uploader $(GO_PREFIX)/cmd/vvgo-uploader


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

.PHONY: images images/vvgo images/object-cache images/page-cache

images/vvgo-builder:
	docker pull $(VVGO_IMAGE_CACHE)-builder || true
	docker build . \
		--file Dockerfile \
		--target builder \
		--cache-from=$(VVGO_IMAGE_CACHE)-builder \
		--build-arg GITHUB_SHA=$GITHUB_SHA \
		--build-arg GITHUB_REF=$GITHUB_REF \
		--tag builder

images/vvgo: images/vvgo-builder
	docker pull $(VVGO_IMAGE_CACHE) || true
	docker build . \
		--file Dockerfile \
		--target vvgo \
		--cache-from=$(VVGO_IMAGE_CACHE)-builder \
		--cache-from=$(VVGO_IMAGE_CACHE) \
		--tag artifact

images/page-cache:
	docker pull $(PAGE_CACHE_IMAGE_CACHE) || true
	docker build infra/object-cache \
		--file infra/object-cache/Dockerfile \
		--cache-from=$(PAGE_CACHE_IMAGE_CACHE) \
		--tag object-cache

images/object-cache:
	docker pull $(OBJECT_CACHE_IMAGE_CACHE) || true
	docker build infra/page-cache \
		--file infra/page-cache/Dockerfile \
		--cache-from=$(OBJECT_CACHE_IMAGE_CACHE) \
		--tag page-cache

images/kv-cache:
	docker pull $(KV_CACHE_IMAGE_CACHE) || true
	docker build infra/kv-cache \
		--file infra/kv-cache/Dockerfile \
		--cache-from=$(KV_CACHE_IMAGE_CACHE) \
		--tag kv-cache

images: images/vvgo images/page-cache images/object-cache images/kv-cache

# Push images

RELEASE_TAG ?= development

.PHONY: push push/object-cache push/page-cache push/vvgo push/vvgo-builder

push/vvgo-builder:
	docker tag builder $(IMAGE_REPO)/$(VVGO_IMAGE_NAME)-builder:$(RELEASE_TAG)
	docker push $(IMAGE_REPO)/$(VVGO_IMAGE_NAME)-builder:$(RELEASE_TAG)

push/vvgo:
	docker tag artifact $(IMAGE_REPO)/$(VVGO_IMAGE_NAME):$(RELEASE_TAG)
	docker push $(IMAGE_REPO)/$(VVGO_IMAGE_NAME):$(RELEASE_TAG)

push/page-cache:
	docker tag page-cache $(IMAGE_REPO)/$(PAGE_CACHE_IMAGE_NAME):$(RELEASE_TAG)
	docker push $(IMAGE_REPO)/$(PAGE_CACHE_IMAGE_NAME):$(RELEASE_TAG)

push/object-cache:
	docker tag object-cache $(IMAGE_REPO)/$(OBJECT_CACHE_IMAGE_NAME):$(RELEASE_TAG)
	docker push $(IMAGE_REPO)/$(OBJECT_CACHE_IMAGE_NAME):$(RELEASE_TAG)

push/kv-cache:
	docker tag kv-cache $(IMAGE_REPO)/$(KV_CACHE_IMAGE_NAME):$(RELEASE_TAG)
	docker push $(IMAGE_REPO)/$(KV_CACHE_IMAGE_NAME):$(RELEASE_TAG)

push: push/vvgo-builder push/vvgo push/page-cache push/object-cache push/kv-cache
