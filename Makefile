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

GITHUB_REPO ?= virtual-vgo/vvgo
IMAGE_REPO ?= docker.pkg.github.com/$(GITHUB_REPO)

.PHONY: images
images: images/vvgo-builder
images: images/vvgo

images/vvgo-builder:
	docker build . \
		--file Dockerfile \
		--target builder \
		--build-arg GITHUB_SHA=$GITHUB_SHA \
		--build-arg GITHUB_REF=$GITHUB_REF \
		--tag vvgo-builder

images/vvgo:
	docker build . \
 	--file Dockerfile \
 	--target vvgo \
 	--tag vvgo

# Deploy images

.PHONY: push
push: push/vvgo-builder
push: push/vvgo

push/%: images/%
	docker tag $* $(IMAGE_REPO)/$*:$(RELEASE_TAG)
	docker push $(IMAGE_REPO)/$*:$(RELEASE_TAG)
