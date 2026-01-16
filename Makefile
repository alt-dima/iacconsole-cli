ifeq ($(strip $(VERSION_STRING)),)
VERSION_STRING := $(shell git rev-parse --short HEAD)
endif

BINDIR    := $(CURDIR)/bin
PLATFORMS := linux/amd64/iacconsole-cli-Linux-x86_64/osusergo*netgo*static_build darwin/amd64/iacconsole-cli-Darwin-x86_64/osusergo*netgo*static_build linux/arm64/iacconsole-cli-Linux-arm64/osusergo*netgo*static_build darwin/arm64/iacconsole-cli-Darwin-arm64/osusergo*netgo*static_build
BUILDCOMMAND := go build -trimpath -ldflags "-s -w -X main.version=${VERSION_STRING}"
temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))
label = $(word 3, $(temp))
tags = $(subst *, ,$(word 4, $(temp)))

UNAME := $(shell uname)
ifeq ($(UNAME), Darwin)
SHACOMMAND := shasum -a 256
else
SHACOMMAND := sha256sum
endif

.DEFAULT_GOAL := build

.PHONY: release
release: $(PLATFORMS)
$(PLATFORMS):
	GOOS=$(os) GOARCH=$(arch) CGO_ENABLED=0 $(BUILDCOMMAND) -tags "$(tags)" -o "bin/$(label)"
	$(SHACOMMAND) "bin/$(label)" > "bin/$(label).sha256"

.PHONY: latest
latest:
	echo ${VERSION_STRING} > bin/latest

.PHONY: build
build:
	$(BUILDCOMMAND) -o ${BINDIR}/iacconsole-cli

.PHONY: dep
dep:
	go mod tidy
