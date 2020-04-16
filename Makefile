# Makefile

GO_PREFIX ?= github.com/virtual-vgo/vvgo

# Build vvgo
.PHONY: vvgo vvgo-uploader # Use go build tools caching
BIN_PATH ?= .
BUILD_FLAGS ?= -v
vvgo:
	go generate ./... && go build -v -o $(BIN_PATH)/$@ $(GO_PREFIX)/$@
vvgo-uploader:
	go generate ./... && go build -v -o $(BIN_PATH)/$@ $(GO_PREFIX)/$@

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
images: images/vvgo images/object-cache images/page-cache images/kv-cache
images/vvgo-builder:
	docker pull $(VVGO_IMAGE_CACHE)-builder || true
	docker build . \
		--file Dockerfile \
		--target builder \
		--cache-from=$(VVGO_IMAGE_CACHE)-builder \
		--build-arg GITHUB_SHA=$GITHUB_SHA \
		--build-arg GITHUB_REF=$GITHUB_REF \
		--tag builder

images/vvgo:
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

# Releases

HARDWARE ?= $(shell uname -m)
RELEASE_TAG ?= $(shell git rev-parse --short HEAD)

.PHONY: releases releases/$(BIN_PATH) releases/$(IMAGE_REPO)
releases: releases/$(BIN_PATH)
releases/$(BIN_PATH): $(BIN_PATH)/vvgo-uploader-$(RELEASE_TAG)-linux-$(HARDWARE)
releases/$(BIN_PATH): $(BIN_PATH)/vvgo-uploader-$(RELEASE_TAG)-darwin-$(HARDWARE)
releases/$(BIN_PATH): $(BIN_PATH)/vvgo-uploader-$(RELEASE_TAG)-windows-$(HARDWARE).exe

$(BIN_PATH)/%-$(RELEASE_TAG)-darwin-$(HARDWARE):
	GOOS=darwin go build -v -o $(BIN_PATH)/$@-$(RELEASE_TAG)-darwin-$(HARDWARE).exe $(GO_PREFIX)/cmd/vvgo-uploader
$(BIN_PATH)/%-$(RELEASE_TAG)-linux-$(HARDWARE):
	GOOS=linux go build -v -o $(BIN_PATH)/$@-$(RELEASE_TAG)-linux-$(HARDWARE) $(GO_PREFIX)/cmd/vvgo-uploader
$(BIN_PATH)/%-$(RELEASE_TAG)-windows-$(HARDWARE).exe:
	GOOS=windows go build -v -o $(BIN_PATH)/$@-$(RELEASE_TAG)-windows-$(HARDWARE).exe $(GO_PREFIX)/cmd/vvgo-uploader

releases: releases/$(IMAGE_REPO)
releases/$(IMAGE_REPO): releases/$(IMAGE_REPO)/$(VVGO_IMAGE_NAME)-builder\:$(RELEASE_TAG)
releases/$(IMAGE_REPO): releases/$(IMAGE_REPO)/$(VVGO_IMAGE_NAME)\:$(RELEASE_TAG)
releases/$(IMAGE_REPO): releases/$(IMAGE_REPO)/$(PAGE_CACHE_IMAGE_NAME)\:$(RELEASE_TAG)
releases/$(IMAGE_REPO): releases/$(IMAGE_REPO)/$(OBJECT_CACHE_IMAGE_NAME)\:$(RELEASE_TAG)
releases/$(IMAGE_REPO): releases/$(IMAGE_REPO)/$(KV_CACHE_IMAGE_NAME)\:$(RELEASE_TAG)

releases/$(IMAGE_REPO)/%: images/$(IMAGE_REPO)/%
	docker tag $@ $(IMAGE_REPO)/$@

# Push images

.PHONY: push push/$(IMAGE_REPO)
push: push/$(IMAGE_REPO)
push/$(IMAGE_REPO): push/$(IMAGE_REPO)/vvgo-builder\:$(RELEASE_TAG)
push/$(IMAGE_REPO): push/$(IMAGE_REPO)/vvgo\:$(RELEASE_TAG)
push/$(IMAGE_REPO): push/$(IMAGE_REPO)/page-cache\:$(RELEASE_TAG)
push/$(IMAGE_REPO): push/$(IMAGE_REPO)/object-cache\:$(RELEASE_TAG)
push/$(IMAGE_REPO): push/$(IMAGE_REPO)/kv-cache\:$(RELEASE_TAG)

push/$(IMAGE_REPO)/%:
	docker push $(IMAGE_REPO)/$@
