MAKEFLAGS += --warn-undefined-variables
CWD := $(abspath $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST))))))
SHELL := bash
.SHELLFLAGS := -euo pipefail -c
.DEFAULT_GOAL := all
APPNAME := "wireguard-cni"

# Build dependencies go in here
BIN=.bin
BUF=./.bin/buf

# Respect $GOBIN if set in environment or via $GOENV file.
GOBIN_DIR ?= $(shell go env GOBIN)
BIN_DIR ?= $(shell go env GOPATH)/bin
export PATH := ${CWD}/${BIN}:$(PATH):$(BIN_DIR)

ifneq ("$(wildcard .makefiles/*.mk)","")
	include .makefiles/*.mk
else
    $(info "no makefiles to load")
endif

.PHONY: all
all: proto mocks test build

.PHONY: proto
proto: buf/lint deps
	@$(BIN)/buf generate

.PHONY: buf/lint
buf/lint: deps
	@$(BIN)/buf lint

.PHONY: deps
deps: ./.bin/buf ./.bin/protoc-gen-go ./.bin/protoc-gen-connect-go ## deps: installs build time dependencies

.PHONY: extra-deps
extra-deps: ./.bin/hc-install ./.bin/nomad ./.bin/vagrant ## extra-deps: installs other tools like nomad and vagrant

.PHONY: mocks
mocks: ./.bin/mockery
	@./.bin/mockery --all --inpackage --log-level info

./.bin/mockery:
	GOBIN=${CWD}/${BIN} go install github.com/vektra/mockery/v2@v2.40.1

./.bin/buf:
	GOBIN=${CWD}/${BIN} go install github.com/bufbuild/buf/cmd/buf@v1.5.0

./.bin/protoc-gen-go:
	GOBIN=${CWD}/${BIN} go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.0

./.bin/protoc-gen-connect-go:
	GOBIN=${CWD}/${BIN} go install connectrpc.com/connect/cmd/protoc-gen-connect-go@v1.12.0

./.bin/hc-install:
	GOBIN=${CWD}/${BIN} go install github.com/hashicorp/hc-install/cmd/hc-install@main

./.bin/nomad: ./.bin/hc-install
	hc-install install -version 1.3.2 -path ./.bin nomad

./.bin/vagrant: ./.bin/hc-install
	hc-install install -version 2.2.19 -path ./.bin vagrant
