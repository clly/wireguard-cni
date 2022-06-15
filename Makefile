MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -euo pipefail -c
.DEFAULT_GOAL := all
APPNAME := "wireguard-cni"

BUF_VERSION := 1.5.0
BIN := .bin
BUF := .bin/buf
export DO_NOT_TRACK=1

BIN_DIR ?= $(shell go env GOPATH)/bin
export PATH := $(PATH):$(BIN_DIR)

ifneq ("$(wildcard .makefiles/*.mk)","")
	include .makefiles/*.mk
else
    $(info "no makefiles to load")
endif

.PHONY: all
all: proto test build

.PHONY: proto
proto: $(BUF) buf/lint
	@buf generate

.PHONY: buf/install
buf/install: ## installs buf
	@curl -sSL "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-$$(uname -s)-$$(uname -m)" \
    -o "${BIN}/buf" && \
  chmod +x "${BIN}/buf"

$(BUF): buf/install

.PHONY: buf/lint
buf/lint: $(BUF)
	@$(BUF) lint
