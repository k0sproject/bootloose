# SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
# SPDX-FileCopyrightText: 2023 bootloose authors
# SPDX-License-Identifier: Apache-2.0
GO_SRCS := $(shell find . -type f -name '*.go' -a ! \( -name 'zz_generated*' -o -name '*_test.go' \))
GO_TESTS := $(shell find . -type f -name '*_test.go')
TAG_NAME = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
GIT_COMMIT = $(shell git rev-parse --short=7 HEAD)
ifdef TAG_NAME
	ENVIRONMENT = production
endif
ENVIRONMENT ?= development
LD_FLAGS = -s -w -X github.com/k0sproject/bootloose/version.Environment=$(ENVIRONMENT) -X github.com/carlmjohnson/versioninfo.Revision=$(GIT_COMMIT) -X github.com/carlmjohnson/versioninfo.Version=$(TAG_NAME)
BUILD_FLAGS = -trimpath -a -tags "netgo,osusergo,static_build" -installsuffix netgo -ldflags "$(LD_FLAGS) -extldflags '-static'"
PREFIX = /usr/local

BIN_PREFIX := bootloose-

all: bootloose

bootloose:
	go build -v $(BUILD_FLAGS) -o bootloose main.go

PLATFORMS := linux-amd64 linux-arm64 linux-arm darwin-amd64 darwin-arm64
bins := $(foreach platform, $(PLATFORMS), bin/$(BIN_PREFIX)$(platform))

$(bins):
	$(eval temp := $(subst -, ,$(subst $(BIN_PREFIX),,$(notdir $@))))
	$(eval OS := $(word 1, $(subst -, ,$(temp))))
	$(eval ARCH := $(word 2, $(subst -, ,$(temp))))
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $@ main.go

bin/%: $(GO_SRCS)
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_FLAGS) -o $@ main.go

bin/sha256sums.txt: $(bins)
	sha256sum -b $(bins) | sed 's|bin/||' > $@

# for use in release markdown
bin/sha256sums.md: bin/sha256sums.txt
	@echo "### SHA256 Checksums" > $@
	@echo >> $@
	@echo "\`\`\`" >> $@
	@cat $< >> $@
	@echo "\`\`\`" >> $@

build-all: $(bins) bin/sha256sums.md

install: bootloose
	install -d $(DESTDIR)$(PREFIX)/bin/
	install -m 755 bootloose $(DESTDIR)$(PREFIX)/bin/

# Build all images
images:
	@$(MAKE) -C images all

# Build a specific image
image-%:
	@$(MAKE) -C images $*

test-unit:
	go test -v . ./pkg/...

# Run tests against all images
test-e2e:
	@$(MAKE) -C tests test-all

# Run tests against a specific image
test-e2e-%:
	@$(MAKE) -C tests test-$*

test: test-unit test-e2e

# List available images
list-images:
	@$(MAKE) -C images list

# Clean up all stamps and other generated files
clean:
	@$(MAKE) -C images clean
	rm -f bootloose bin/*

# Phony targets
.PHONY: install images image-% test-unit test-e2e test-e2e-% list-images clean build-all
