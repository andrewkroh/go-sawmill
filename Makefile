.PHONY: all
all: fmt generate-processors generate-readme test

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
	@go run github.com/elastic/go-licenser@latest
	@echo go-fumpt
	@go run mvdan.cc/gofumpt@latest -w --extra $(shell find . -name '*.go')

.PHONY: test
test:
	go test ./...
