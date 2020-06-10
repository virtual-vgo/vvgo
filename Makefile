# Makefile

PATH := $(GOROOT)/bin:$(PATH)
GO_PREFIX ?= github.com/virtual-vgo/vvgo

# Quick build cmds
.PHONY: vvgo vvgo-uploader # Use go build tools caching
BIN_PATH ?= .
BUILD_FLAGS ?= -v

default: vvgo

all: node_modules test releases images

vvgo:
	go generate ./... && go build -o $(BIN_PATH)/$@ $(GO_PREFIX)/cmd/vvgo
vvgo-uploader:
	go generate ./... && go build -o $(BIN_PATH)/$@ $(GO_PREFIX)/cmd/vvgo-uploader

# Download external dependencies (yarn)
.PHONY: node_modules
node_modules:
	yarn install

# Generate code
generate: cmd/vvgo/info.go cmd/vvgo-uploader/info.go
cmd/vvgo/info.go:
	go generate $(GO_PREFIX)/cmd/vvgo
cmd/vvgo-uploader/info.go:
	go generate $(GO_PREFIX)/cmd/vvgo-uploader

# Run tests
.PHONY: fmt vet test
fmt:
	$(GOFMT) -d .
vet:
	go vet $(GO_PREFIX)/...
TEST_FLAGS ?= -race
test: generate fmt vet
	go test $(TEST_FLAGS) $(GO_PREFIX)/...

# Make releases

HARDWARE ?= $(shell uname -m)
RELEASE_TAG ?= $(shell git rev-parse --short HEAD)

.PHONY: releases releases/$(IMAGE_REPO)
releases: $(BIN_PATH)/vvgo-$(RELEASE_TAG)-linux-$(HARDWARE)
releases: $(BIN_PATH)/vvgo-$(RELEASE_TAG)-darwin-$(HARDWARE)
releases: $(BIN_PATH)/vvgo-$(RELEASE_TAG)-windows-$(HARDWARE).exe
releases: $(BIN_PATH)/vvgo-uploader-$(RELEASE_TAG)-linux-$(HARDWARE)
releases: $(BIN_PATH)/vvgo-uploader-$(RELEASE_TAG)-darwin-$(HARDWARE)
releases: $(BIN_PATH)/vvgo-uploader-$(RELEASE_TAG)-windows-$(HARDWARE).exe
releases: $(BIN_PATH)/vvgo-uploader-$(RELEASE_TAG)-windows-$(HARDWARE).exe.zip

$(BIN_PATH)/%-$(RELEASE_TAG)-darwin-$(HARDWARE): generate
	GOOS=darwin go build -o $@ $(GO_PREFIX)/cmd/$*
$(BIN_PATH)/%-$(RELEASE_TAG)-linux-$(HARDWARE): generate
	GOOS=linux go build -o $@ $(GO_PREFIX)/cmd/$*
$(BIN_PATH)/%-$(RELEASE_TAG)-windows-$(HARDWARE).exe: generate
	GOOS=windows go build -o $@ $(GO_PREFIX)/cmd/$*

%.tar.gz: %
	tar czf $@ $^

%.zip: %
	zip $@ $^

# Build images

GITHUB_REPO ?= virtual-vgo/vvgo
IMAGE_REPO ?= docker.pkg.github.com/$(GITHUB_REPO)

.PHONY: images
images: images/vvgo-builder
images: images/vvgo

images/vvgo:
	docker build . \
		--file Dockerfile \
		--build-arg GITHUB_SHA=$GITHUB_SHA \
		--build-arg GITHUB_REF=$GITHUB_REF \
		--tag vvgo

# Deploy images

.PHONY: push
push: push/vvgo

push/%: images/%
	docker tag $* $(IMAGE_REPO)/$*:$(RELEASE_TAG)
	docker push $(IMAGE_REPO)/$*:$(RELEASE_TAG)
