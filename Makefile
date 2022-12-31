GOFUMPT := go run mvdan.cc/gofumpt@latest
GOLICENSER := go run github.com/elastic/go-licenser@latest
GORELEASER := go run github.com/goreleaser/goreleaser@latest

.PHONY: all
all: fmt generate test build examples

.PHONY: generate
generate: generate-processors generate-readme

.PHONY: generate-readme
generate-readme:
	go run docs/generate/generate.go -p pkg/processor/processors.yml -o docs/README.md

.PHONY: generate-processors
generate-processors:
	go generate ./pkg/processor

.PHONY: fmt
fmt:
	go mod tidy
	@echo go-licenser
	@${GOLICENSER}
	@echo go-fumpt
	@${GOFUMPT} -w --extra $(shell find . -name '*.go')

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	${GORELEASER} build --snapshot --rm-dist

.PHONY: examples
examples:
	$(MAKE) -C examples/c
	$(MAKE) -C examples/wasm build

