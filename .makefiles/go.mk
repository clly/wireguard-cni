.PHONY: deps

BINS=$(wildcard cmd/*)
deps: ## download go modules
	go mod download

.PHONY: vendor
vendor: ## download and vendor dependencies
	go mod vendor

.PHONY: fmt
fmt: lint/check ## ensure consistent code style
	golangci-lint run --fix > /dev/null 2>&1 || true
	go mod tidy

.PHONY: lint/check
lint/check:  lint/install
	@if ! golangci-lint --version > /dev/null 2>&1; then \
		echo -e "golangci-lint is not installed: run \`make lint/install\` or install it from https://golangci-lint.run"; \
		exit 1; \
	fi

.PHONY: lint/install
lint/install: ## installs golangci-lint to the go bin dir
	@if ! golangci-lint --version > /dev/null 2>&1; then \
		echo "Installing golangci-lint"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(BIN_DIR) v1.46.2; \
	fi

.PHONY: lint
lint: lint/check ## run golangci-lint
	@if [[ -f .golangci-lint.yaml ]]; then golangci-lint --config .golangci-lint.yaml run ; else golangci-lint --config .makefiles/config/golangci-lint.yaml run; fi

.PHONY: test
test: lint ## run go tests
	go test ./... -race -timeout 1m

.PHONY: build
build: ## compile and build artifact
	@for i in cmd/*; do \
		echo "building $$i"; \
		go build -o bin/$$i ./$$i; \
	done

.PHONY: docker/build
docker/build: ## compile and build binaries in a docker contianer
	@docker run --rm \
	-v $$PWD:/app --workdir /app \
	--user $$(id -u) \
	-e GOCACHE=/tmp \
	golang:1.18  make build


.PHONY: build/cmd
build/cmd:
	go build ./cmd/...
