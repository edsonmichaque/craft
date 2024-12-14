package main

const makefileTemplate = `BINARIES ?= {{range .Binaries}}{{.}} {{end}}

VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT  ?= $(shell git rev-parse HEAD)
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -X '{{.ModulePrefix}}/pkg/version.Version=${VERSION}' \
           -X '{{.ModulePrefix}}/pkg/version.GitCommit=${COMMIT}' \
           -X '{{.ModulePrefix}}/pkg/version.BuildTime=${DATE}'

.PHONY: all
all: deps build

.PHONY: deps
deps:
	./scripts/ci/tasks/dependencies.sh

.PHONY: build
build:
	./scripts/ci/tasks/build.sh

.PHONY: test
test:
	./scripts/ci/tasks/test.sh

.PHONY: lint
lint:
	./scripts/ci/tasks/lint.sh

.PHONY: docker
docker:
	./scripts/ci/tasks/docker.sh

.PHONY: proto
proto:
	./scripts/ci/tasks/proto.sh

.PHONY: release
release:
	./scripts/ci/tasks/release.sh $(version)

.PHONY: package
package:
	./scripts/ci/tasks/package.sh

.PHONY: dev-setup
dev-setup:
	./scripts/ci/utils/setup-dev.sh

.PHONY: clean
clean:
	./scripts/ci/utils/cleanup.sh

.PHONY: deep-clean
deep-clean:
	./scripts/ci/utils/cleanup.sh --deep

.PHONY: health-check
health-check:
	./scripts/ci/utils/health-check.sh
`
