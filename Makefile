MAKEFLAGS += --warn-undefined-variables
CWD := $(abspath $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST))))))
SHELL := bash
.SHELLFLAGS := -euo pipefail -c
.DEFAULT_GOAL := all
APPNAME := "wireguard-cni"

BUF_VERSION := 1.5.0
BIN := .bin
BUF := .bin/buf
export DO_NOT_TRACK=1

BIN_DIR ?= $(shell go env GOPATH)/bin
export PATH := ${CWD}/${BIN}:$(PATH):$(BIN_DIR)

ifneq ("$(wildcard .makefiles/*.mk)","")
	include .makefiles/*.mk
else
    $(info "no makefiles to load")
endif

.PHONY: all
all: proto test build

.PHONY: proto
proto: $(BUF) buf/lint deps
	@$(BUF) generate

.PHONY: buf/install
buf/install: ## installs buf
	curl -sSL "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-$$(uname -s)-$$(uname -m)" \
    -o "${BIN}/buf" && \
  chmod +x "${BIN}/buf"

$(BUF):
	make buf/install

.PHONY: buf/lint
buf/lint: $(BUF)
	@$(BUF) lint

.PHONY: deps
deps: ./.bin/protoc-gen-go ./.bin/protoc-gen-connect-go ## deps installs build time dependencies

.PHONY: extra-deps
extra-deps: ./.bin/hc-install ./.bin/nomad ./.bin/vagrant ## extra deps installs helpful dependencies like nomad and vagrant

./.bin/protoc-gen-go:
	GOBIN=${CWD}/${BIN} go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.0

./.bin/protoc-gen-connect-go:
	GOBIN=${CWD}/${BIN} go install github.com/bufbuild/connect-go/cmd/protoc-gen-connect-go@v0.2.0

./.bin/hc-install:
	GOBIN=${CWD}/${BIN} go install github.com/hashicorp/hc-install/cmd/hc-install@main

./.bin/nomad: ./.bin/hc-install
	hc-install install -version 1.3.2 -path ./.bin nomad

./.bin/vagrant: ./.bin/hc-install
	hc-install install -version 2.2.19 -path ./.bin vagrant