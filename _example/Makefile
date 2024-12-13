BINARIES ?= craftd craftctl 

VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT  ?= $(shell git rev-parse HEAD)
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -X 'github.com/edsonmichaque/craft/pkg/version.Version=${VERSION}' \
           -X 'github.com/edsonmichaque/craft/pkg/version.GitCommit=${COMMIT}' \
           -X 'github.com/edsonmichaque/craft/pkg/version.BuildTime=${DATE}'

.PHONY: all
all: deps build

.PHONY: deps
deps:
	go mod download

.PHONY: build
build:
	@for binary in $(BINARIES); do \
		echo "Building $$binary..." ; \
		go build -ldflags "$(LDFLAGS)" -o bin/$$binary ./cmd/$$binary ; \
	done

.PHONY: test
test:
	go test -v -race -cover ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: docker
docker:
	@for binary in $(BINARIES); do \
		echo "Building Docker image for $$binary..." ; \
		docker build \
			--build-arg VERSION=$(VERSION) \
			--build-arg COMMIT=$(COMMIT) \
			--build-arg BUILD_TIME=$(DATE) \
			-t craft-$$binary:$(VERSION) \
			-f docker/$$binary.Dockerfile . ; \
	done

.PHONY: proto
proto:
	@if [ -d "proto" ]; then \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			proto/*.proto ; \
	fi

.PHONY: clean
clean:
	rm -rf bin/
	rm -rf dist/
