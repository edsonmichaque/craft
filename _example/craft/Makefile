# Makefile for CI Scripts

.PHONY: all test build clean docker package ci help

# Environment variables
export PROJECT_ROOT := $(shell pwd)
export CI_SCRIPTS := $(PROJECT_ROOT)/scripts/ci

# Default target
all: test build

# Help message
help:
	@echo "Available targets:"
	@echo "  test       - Run tests"
	@echo "  build      - Build project"
	@echo "  clean      - Clean build artifacts"
	@echo "  docker     - Build Docker images"
	@echo "  package    - Create distribution packages"
	@echo "  ci         - Run full CI pipeline"
	@echo "  ci-test    - Test CI configurations locally"
	@echo "  setup-dev  - Setup development environment"

# Main targets
test:
	@$(CI_SCRIPTS)/test $(ARGS)

test-watch:
	@$(CI_SCRIPTS)/test --watch $(ARGS)

test-coverage:
	@$(CI_SCRIPTS)/test coverage $(ARGS)

build:
	@$(CI_SCRIPTS)/build $(ARGS)

build-debug:
	@DEBUG=1 $(CI_SCRIPTS)/build $(ARGS)

clean:
	@$(CI_SCRIPTS)/utils/cleanup.sh $(ARGS)

docker:
	@$(CI_SCRIPTS)/build docker $(ARGS)

package:
	@$(CI_SCRIPTS)/build package $(ARGS)

ci:
	@$(CI_SCRIPTS)/ci $(ARGS)

# Development targets
setup-dev:
	@$(CI_SCRIPTS)/setup-dev.sh

lint:
	@$(CI_SCRIPTS)/test lint $(ARGS)

generate:
	@$(CI_SCRIPTS)/tasks/generate.sh $(ARGS)

# CI testing targets
ci-test:
	@$(CI_SCRIPTS)/utils/ci-tester.sh $(ARGS)

ci-test-github:
	@$(CI_SCRIPTS)/utils/ci-tester.sh github $(ARGS)

ci-test-gitlab:
	@$(CI_SCRIPTS)/utils/ci-tester.sh gitlab $(ARGS)

# Database targets
db-start:
	@$(CI_SCRIPTS)/utils/db.sh start $(ARGS)

db-migrate:
	@$(CI_SCRIPTS)/utils/db.sh migrate $(ARGS)

db-seed:
	@$(CI_SCRIPTS)/utils/db.sh seed $(ARGS)

# Release targets
release:
	@$(CI_SCRIPTS)/tasks/release.sh $(ARGS)

release-rc:
	@$(CI_SCRIPTS)/tasks/release.sh --rc $(ARGS)

release-hotfix:
	@$(CI_SCRIPTS)/tasks/release.sh --hotfix $(ARGS)