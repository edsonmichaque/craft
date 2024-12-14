package main

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//go:embed readme.adoc.tmpl
var readmeTemplate string

const jenkinsTemplate = `
pipeline {
    agent any

    environment {
        GO_VERSION = '1.21'
        PROJECT_NAME = 'your-project-name'
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Set up Go') {
            steps {
                script {
                    sh 'go version'
                    sh "go mod download"
                }
            }
        }

        stage('Lint') {
            steps {
                script {
                    sh 'make lint'
                }
            }
        }

        stage('Test') {
            steps {
                script {
                    sh 'make test'
                }
            }
        }

        stage('Build') {
            steps {
                script {
                    sh 'make build'
                }
            }
        }

        stage('Docker Build and Push') {
            when {
                expression { return env.BRANCH_NAME == 'main' || env.BRANCH_NAME.startsWith('release/') }
            }
            steps {
                script {
                    sh 'make docker'
                }
            }
        }

        stage('Release') {
            when {
                expression { return env.TAG_NAME != null }
            }
            steps {
                script {
                    sh 'make package'
                }
            }
        }
    }

    post {
        always {
            archiveArtifacts artifacts: 'bin/**', allowEmptyArchive: true
            junit 'coverage/coverage.xml'
        }
        success {
            echo 'Pipeline completed successfully!'
        }
        failure {
            echo 'Pipeline failed!'
        }
    }
}`

const githubWorkflowTemplate = `name: CI

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '{{.GoVersion}}'
          
      - name: Run Tests
        run: make test
        env:
          LOG_LEVEL: DEBUG
          
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '{{.GoVersion}}'
          
      - name: Run Linters
        run: make lint

  build:
    needs: [test, lint]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '{{.GoVersion}}'
          
      - name: Build Project
        run: make build
        
      - name: Upload Artifacts
        uses: actions/upload-artifact@v4
        with:
          name: binaries
          path: bin/

  docker:
    if: startsWith(github.ref, 'refs/tags/v')
    needs: [build]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
        
      - name: Login to Docker Hub
        uses: docker/login-action@v3
        with:
          username: {{ "{{" }} secrets.DOCKERHUB_USERNAME {{ "}}" }}
          password: {{ "{{" }} secrets.DOCKERHUB_TOKEN {{ "}}" }}
          
      - name: Build and Push Docker Images
        run: make docker
        env:
          DOCKER_PUSH: true

  release:
    if: startsWith(github.ref, 'refs/tags/v')
    needs: [docker]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Download Artifacts
        uses: actions/download-artifact@v4
        with:
          name: binaries
          path: bin/
          
      - name: Create Release
        run: make package
        env:
          GITHUB_TOKEN: {{ "{{" }} secrets.GITHUB_TOKEN {{ "}}" }}
`

const gitlabCITemplate = `include:
  - local: '.gitlab/ci/test.yml'
  - local: '.gitlab/ci/build.yml'
  - local: '.gitlab/ci/release.yml'

image: golang:{{.GoVersion}}

variables:
  GO111MODULE: "on"
  CGO_ENABLED: "0"
  DOCKER_HOST: tcp://docker:2375
  DOCKER_DRIVER: overlay2

services:
  - docker:dind

stages:
  - test
  - build
  - package
  - deploy

before_script:
  - go mod download
`

const gitlabTestTemplate = `lint:
  stage: test
  script:
    - make lint

test:
  stage: test
  script:
    - make test
  coverage: '/coverage: \d+\.\d+/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage/coverage.xml
`

const gitlabBuildTemplate = `build:
  stage: build
  script:
    - make build
  artifacts:
    paths:
      - bin/

docker:
  stage: package
  script:
    - make docker
  only:
    - tags
`

const gitlabReleaseTemplate = `release:
  stage: deploy
  script:
    - make package
  only:
    - tags
`

const travisCITemplate = `language: go

go:
  - '{{.GoVersion}}'

services:
  - docker

env:
  global:
    - GO111MODULE=on
    - CGO_ENABLED=0

before_install:
  - curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
  - sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
  - sudo apt-get update
  - sudo apt-get -y install docker-ce

install:
  - go mod download

jobs:
  include:
    - stage: test
      name: "Lint"
      script: make lint
      
    - stage: test
      name: "Unit Tests"
      script: make test
      after_success:
        - bash <(curl -s https://codecov.io/bash)
      
    - stage: build
      name: "Build Binaries"
      script: make build
      
    - stage: package
      name: "Build Docker Images"
      if: tag IS present
      script: make docker
      
    - stage: deploy
      name: "Create Release"
      if: tag IS present
      script: make package

stages:
  - test
  - build
  - package
  - name: deploy
    if: tag IS present

cache:
  directories:
    - $HOME/.cache/go-build
    - $HOME/gopath/pkg/mod
`

const circleCITemplate = `version: 2.1

orbs:
  go: circleci/go@1.9
  docker: circleci/docker@2.4

executors:
  golang:
    docker:
      - image: cimg/go:{{.GoVersion}}
    environment:
      GO111MODULE: "on"
      CGO_ENABLED: "0"

jobs:
  lint:
    executor: golang
    steps:
      - checkout
      - go/mod-download
      - run:
          name: Run Linters
          command: make lint

  test:
    executor: golang
    steps:
      - checkout
      - go/mod-download
      - run:
          name: Run Tests
          command: make test
      - store_test_results:
          path: coverage
      - store_artifacts:
          path: coverage

  build:
    executor: golang
    steps:
      - checkout
      - go/mod-download
      - run:
          name: Build Project
          command: make build
      - persist_to_workspace:
          root: .
          paths:
            - bin

  docker:
    executor: docker/docker
    steps:
      - checkout
      - setup_remote_docker
      - attach_workspace:
          at: .
      - docker/build:
          image: $CIRCLE_PROJECT_USERNAME/$CIRCLE_PROJECT_REPONAME
          tag: ${CIRCLE_TAG:-latest}
      - docker/push:
          image: $CIRCLE_PROJECT_USERNAME/$CIRCLE_PROJECT_REPONAME
          tag: ${CIRCLE_TAG:-latest}
          
  release:
    executor: golang
    steps:
      - checkout
      - attach_workspace:
          at: .
      - run:
          name: Create Release
          command: make package

workflows:
  version: 2
  build-test-deploy:
    jobs:
      - lint:
          filters:
            tags:
              only: /^v.*/
      - test:
          filters:
            tags:
              only: /^v.*/
      - build:
          requires:
            - lint
            - test
          filters:
            tags:
              only: /^v.*/
      - docker:
          requires:
            - build
          filters:
            tags:
              only: /^v.*/
            branches:
              ignore: /.*/
      - release:
          requires:
            - docker
          filters:
            tags:
              only: /^v.*/
            branches:
              ignore: /.*/
`

func generateCIScripts1(projectPath string, cfg Config) error {
	ciDir := filepath.Join(projectPath, "scripts", "ci")

	// Create subdirectories for better organization
	dirs := []string{
		"lib",   // Shared libraries
		"tasks", // Individual task scripts
		"utils", // Utility scripts
		"hooks", // Git hooks
		"env",   // Environment configurations
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(ciDir, dir), 0755); err != nil {
			return fmt.Errorf("failed to create CI directory %s: %w", dir, err)
		}
	}

	// Generate CI configuration files
	ciConfigs := map[string]string{
		".github/workflows/ci.yml": githubWorkflowTemplate,
		".gitlab-ci.yml":           gitlabCITemplate,
		".gitlab/ci/test.yml":      gitlabTestTemplate,
		".gitlab/ci/build.yml":     gitlabBuildTemplate,
		".gitlab/ci/release.yml":   gitlabReleaseTemplate,
		".travis.yml":              travisCITemplate,
		".circleci/config.yml":     circleCITemplate,
		"Jenkinsfile":              jenkinsTemplate,
	}

	for filename, content := range ciConfigs {
		filepath := path.Join(projectPath, filename)

		// Create parent directory if it doesn't exist
		if filename != ".gitlab-ci.yml" && filename != ".travis.yml" && filename != "Jenkinsfile" {
			if err := os.MkdirAll(path.Dir(filepath), 0755); err != nil {
				return fmt.Errorf("failed to create directory for %s: %w", filename, err)
			}
		}

		if err := generateFileFromTemplate(filepath, content, cfg); err != nil {
			return fmt.Errorf("failed to generate CI config %s: %w", filename, err)
		}
	}

	// Generate script files
	scripts := map[string]string{
		// Core library files
		"lib/common.sh":  commonLibTemplate,
		"lib/logger.sh":  loggerLibTemplate,
		"lib/docker.sh":  dockerLibTemplate,
		"lib/git.sh":     gitLibTemplate,
		"lib/version.sh": versionLibTemplate,

		// Task scripts
		"tasks/build.sh":        buildTaskTemplate,
		"tasks/test.sh":         testTaskTemplate,
		"tasks/lint.sh":         lintTaskTemplate,
		"tasks/release.sh":      releaseTaskTemplate,
		"tasks/docker.sh":       dockerTaskTemplate,
		"tasks/proto.sh":        protoTaskTemplate,
		"tasks/dependencies.sh": dependenciesTaskTemplate,
		"tasks/package.sh":      packageTaskTemplate,

		// Utility scripts
		"utils/health-check.sh": healthCheckTemplate,
		"utils/cleanup.sh":      cleanupTemplate,
		"utils/setup-dev.sh":    setupDevTemplate,

		// Main entry points
		"build":       mainBuildTemplate,
		"test":        mainTestTemplate,
		"ci":          mainCITemplate,
		"README.adoc": readmeTemplate,

		"tests/lib/test_helper.bash": testHelperTemplate,
		"tests/lib/common_test.bats": commonTestTemplate,
		"tests/lib/logger_test.bats": loggerTestTemplate,
		"tests/lib/docker_test.bats": dockerTestTemplate,
	}

	for filename, content := range scripts {
		filepath := path.Join(ciDir, filename)

		// Ensure parent directory exists
		if err := os.MkdirAll(path.Dir(filepath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", filename, err)
		}

		if err := generateFileFromTemplate(filepath, content, cfg); err != nil {
			return fmt.Errorf("failed to generate CI script %s: %w", filename, err)
		}

		// Make the script executable
		if err := os.Chmod(filepath, 0755); err != nil {
			return fmt.Errorf("failed to make %s executable: %w", filename, err)
		}
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func generateCIScripts(projectPath string, cfg Config) error {
	// Check if any CI services are specified
	if len(cfg.CIServices) == 0 {
		fmt.Println("No CI services specified, skipping CI script generation.")
		return nil
	}

	ciDir := filepath.Join(projectPath, "scripts", "ci")

	// Create subdirectories for better organization
	dirs := []string{
		"lib",   // Shared libraries
		"tasks", // Individual task scripts
		"utils", // Utility scripts
		"hooks", // Git hooks
		"env",   // Environment configurations
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(ciDir, dir), 0755); err != nil {
			return fmt.Errorf("failed to create CI directory %s: %w", dir, err)
		}
	}

	// Generate build system files
	buildFiles := map[string]string{
		"Makefile":     makefileTemplate2,
		"Taskfile.yml": taskfileTemplate,
		"justfile":     justfileTemplate,
		"magefile.go":  magefileTemplate,
		"Rakefile":     rakefileTemplate,
		"tasks.py":     invokeTemplate,
		"dagger.cue":   daggerTemplate,
	}

	for filename, content := range buildFiles {
		filepath := path.Join(projectPath, filename)
		if err := generateFileFromTemplate(filepath, content, cfg); err != nil {
			return fmt.Errorf("failed to generate build file %s: %w", filename, err)
		}
	}

	// Generate CI configuration files
	ciConfigs := map[string]string{
		".github/workflows/ci.yml": githubWorkflowTemplate,
		".gitlab-ci.yml":           gitlabCITemplate,
		".gitlab/ci/test.yml":      gitlabTestTemplate,
		".gitlab/ci/build.yml":     gitlabBuildTemplate,
		".gitlab/ci/release.yml":   gitlabReleaseTemplate,
		".travis.yml":              travisCITemplate,
		".circleci/config.yml":     circleCITemplate,
		"Jenkinsfile":              jenkinsTemplate,
	}
	// Map CI service names to their config file patterns
	ciServicePatterns := map[string]func(s string) bool{
		"github":   func(s string) bool { return strings.Contains(s, "github") },
		"gitlab":   func(s string) bool { return strings.Contains(s, "gitlab") },
		"circleci": func(s string) bool { return strings.Contains(s, "circleci") },
		"travis":   func(s string) bool { return strings.Contains(s, "travis") },
		"jenkins":  func(s string) bool { return strings.Contains(s, "Jenkinsfile") },
	}

	// Only generate configs for enabled CI services
	for filename, content := range ciConfigs {
		// Check if this config file matches any enabled CI services
		shouldGenerate := false
		for service, matcher := range ciServicePatterns {
			if matcher(filename) {
				for _, enabledService := range cfg.CIServices {
					if service == enabledService {
						shouldGenerate = true
						break
					}
				}
				break
			}
		}

		if !shouldGenerate {
			continue
		}

		filepath := path.Join(projectPath, filename)

		// Create parent directory if it doesn't exist
		if filename != ".gitlab-ci.yml" && filename != ".travis.yml" && filename != "Jenkinsfile" {
			if err := os.MkdirAll(path.Dir(filepath), 0755); err != nil {
				return fmt.Errorf("failed to create directory for %s: %w", filename, err)
			}
		}

		if err := generateFileFromTemplate(filepath, content, cfg); err != nil {
			return fmt.Errorf("failed to generate CI config %s: %w", filename, err)
		}
	}
	// Generate script files
	scripts := map[string]string{
		// Core library files
		"lib/common.sh":  commonLibTemplate,
		"lib/logger.sh":  loggerLibTemplate,
		"lib/docker.sh":  dockerLibTemplate,
		"lib/git.sh":     gitLibTemplate,
		"lib/version.sh": versionLibTemplate,

		// Task scripts
		"tasks/build.sh":        buildTaskTemplate,
		"tasks/test.sh":         testTaskTemplate,
		"tasks/lint.sh":         lintTaskTemplate,
		"tasks/docker.sh":       dockerTaskTemplate,
		"tasks/release.sh":      releaseTaskTemplate,
		"tasks/proto.sh":        protoTaskTemplate,
		"tasks/dependencies.sh": dependenciesTaskTemplate,
		"tasks/package.sh":      packageTaskTemplate,
		//"tasks/generate.sh":     generateTaskTemplate,

		// Utility scripts
		//"utils/cleanup.sh":       cleanupUtilTemplate,
		"utils/health-check.sh": healthCheckTemplate,
		"utils/setup-dev.sh":    setupDevTemplate,
		//"utils/db.sh":            dbUtilTemplate,
		//"utils/ci-tester.sh":     ciTesterTemplate,
		//"utils/install-hooks.sh": installHooksTemplate,

		// Git hooks
		//"hooks/pre-commit": preCommitHookTemplate,
		//"hooks/pre-push":   prePushHookTemplate,

		// Main CI scripts
		"build":       mainBuildTemplate,
		"test":        mainTestTemplate,
		"ci":          mainCITemplate,
		"README.adoc": readmeTemplate,

		// "build": buildMainTemplate,
		// "test":  testMainTemplate,
		// "ci":    ciMainTemplate,
	}

	for filename, content := range scripts {
		filepath := path.Join(ciDir, filename)

		// Create parent directory if it doesn't exist
		if err := os.MkdirAll(path.Dir(filepath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", filename, err)
		}

		if err := generateFileFromTemplate(filepath, content, cfg); err != nil {
			return fmt.Errorf("failed to generate CI script %s: %w", filename, err)
		}

		// Make the script executable
		if err := os.Chmod(filepath, 0755); err != nil {
			return fmt.Errorf("failed to make %s executable: %w", filename, err)
		}
	}

	return nil
}

const commonLibTemplate = `#!/usr/bin/env bash
# Common library for scripts
# Provides core functionality used across all scripts

set -euo pipefail
IFS=$'\n\t'

# Script location
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Import other libraries
source "${SCRIPT_DIR}/logger.sh"
source "${SCRIPT_DIR}/version.sh"

# Global variables
export CI_COMMIT_SHA="${CI_COMMIT_SHA:-$(git rev-parse HEAD)}"
export CI_COMMIT_SHORT_SHA="${CI_COMMIT_SHORT_SHA:-$(git rev-parse --short HEAD)}"
export CI_COMMIT_BRANCH="${CI_COMMIT_BRANCH:-$(git rev-parse --abbrev-ref HEAD)}"
export CI_COMMIT_TAG="${CI_COMMIT_TAG:-$(git describe --tags --exact-match 2>/dev/null || echo '')}"
export CI_PROJECT_NAME="{{.ProjectName}}"

# Environment detection
is_ci() {
    [[ -n "${CI:-}" ]]
}

is_debug() {
    [[ -n "${DEBUG:-}" ]]
}

is_dry_run() {
    [[ -n "${DRY_RUN:-}" ]]
}

# Error handling
trap 'error_handler $?' ERR

error_handler() {
    local exit_code=$1
    log_error "Error occurred in $(caller) with exit code $exit_code"
    exit "$exit_code"
}

# Utility functions
retry() {
    local retries=${1:-3}
    local wait=${2:-5}
    local cmd=${@:3}
    local retry_count=0

    until [[ $retry_count -ge $retries ]]; do
        if eval "$cmd"; then
            return 0
        fi
        retry_count=$((retry_count + 1))
        log_warn "Command failed, attempt $retry_count/$retries"
        sleep "$wait"
    done
    return 1
}

ensure_command() {
    local cmd=$1
    if ! command -v "$cmd" >/dev/null 2>&1; then
        log_error "Required command not found: $cmd"
        exit 1
    fi
}

# Configuration management
load_env() {
    local env=${1:-development}
    local env_file="${PROJECT_ROOT}/scripts/env/${env}.env"
    local local_env_file="${PROJECT_ROOT}/.env"
    
    if [[ -f "$local_env_file" ]]; then
        log_info "Loading local environment from $local_env_file"
        set -o allexport
        source "$local_env_file"
        set +o allexport
    elif [[ -f "$env_file" ]]; then
        log_info "Loading environment from $env_file"
        set -o allexport
        source "$env_file"
        set +o allexport
    else
        log_warn "Environment file not found: $env_file"
    fi
}

# Cleanup handling
cleanup() {
    local exit_code=$?
    log_info "Cleaning up..."
    exit "$exit_code"
}
trap cleanup EXIT
`

const loggerLibTemplate = `#!/usr/bin/env bash
# Logger library for scripts
# Provides standardized logging functionality

# Colors
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly PURPLE='\033[0;35m'
readonly CYAN='\033[0;36m'
readonly GRAY='\033[0;90m'
readonly NC='\033[0m'

# Log levels
declare -A LOG_LEVELS=( 
    ["DEBUG"]=0
    ["INFO"]=1
    ["WARN"]=2
    ["ERROR"]=3
    ["FATAL"]=4
)
LOG_LEVEL=${LOG_LEVEL:-INFO}

# Logging functions
log() {
    local level=$1
    local message=$2
    local color=$3
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    if [[ ${LOG_LEVELS[$level]} -ge ${LOG_LEVELS[$LOG_LEVEL]} ]]; then
        printf "%b%s [%b%s%b] %s%b\n" \
            "$GRAY" "$timestamp" \
            "$color" "$level" "$GRAY" \
            "$message" "$NC" >&2
    fi
}

log_debug() { log "DEBUG" "$1" "$GRAY"; }
log_info() { log "INFO" "$1" "$GREEN"; }
log_warn() { log "WARN" "$1" "$YELLOW"; }
log_error() { log "ERROR" "$1" "$RED"; }
log_fatal() { log "FATAL" "$1" "$RED"; exit 1; }

# Progress indicators
spinner() {
    local pid=$1
    local delay=0.1
    local spinstr='|/-\'
    while ps -p "$pid" > /dev/null; do
        local temp=${spinstr#?}
        printf " [%c]  " "$spinstr"
        local spinstr=$temp${spinstr%"$temp"}
        sleep $delay
        printf "\b\b\b\b\b\b"
    done
    printf "    \b\b\b\b"
}

progress_bar() {
    local current=$1
    local total=$2
    local width=${3:-50}
    local percentage=$((current * 100 / total))
    local completed=$((width * current / total))
    local remaining=$((width - completed))

    printf "\rProgress: ["
    printf "%${completed}s" | tr ' ' '='
    printf "%${remaining}s" | tr ' ' ' '
    printf "] %d%%" "$percentage"
}
`
const dockerLibTemplate = `#!/usr/bin/env bash
# Docker utility functions

source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

# Docker build with caching and multi-stage optimization
docker_build() {
    local image=$1
    local dockerfile=$2
    local context=${3:-.}
    local cache_from=""
    local build_args=()
    local platforms=${DOCKER_PLATFORMS:-"linux/amd64,linux/arm64"}

    # Add build arguments
    build_args+=(--build-arg "VERSION=${CI_COMMIT_TAG:-dev}")
    build_args+=(--build-arg "COMMIT=${CI_COMMIT_SHA}")
    build_args+=(--build-arg "BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")")

    # Use cache from previous builds if available
    if [[ -n "${CI_REGISTRY_IMAGE:-}" ]]; then
        cache_from="--cache-from ${CI_REGISTRY_IMAGE}:${CI_COMMIT_BRANCH:-main}"
    fi

    # Ensure buildx is available and create builder if needed
    if ! docker buildx inspect multiarch >/dev/null 2>&1; then
        log_info "Creating multiarch builder"
        docker buildx create --name multiarch --driver docker-container --use
    fi

    log_info "Building Docker image: $image for platforms: $platforms"
    docker buildx build \
        "${build_args[@]}" \
        $cache_from \
        --platform "$platforms" \
        --push="${DOCKER_PUSH:-false}" \
        -t "$image" \
        -f "$dockerfile" \
        "$context"
}

# Push image with retries and fallback tags
docker_push() {
    local image=$1
    local registry=${2:-}
    local retries=3

    if [[ -n "$registry" ]]; then
        image="${registry}/${image}"
    fi

    log_info "Pushing Docker image: $image"
    retry "$retries" 5 docker push "$image"

    # Tag and push additional tags if this is a release
    if [[ -n "${CI_COMMIT_TAG:-}" ]]; then
        local version_tag="${image}:${CI_COMMIT_TAG}"
        local latest_tag="${image}:latest"
        
        docker tag "$image" "$version_tag"
        docker tag "$image" "$latest_tag"
        
        retry "$retries" 5 docker push "$version_tag"
        retry "$retries" 5 docker push "$latest_tag"
    fi
}

# Clean up Docker resources with configurable options
docker_cleanup() {
    local all=${1:-false}
    local age=${2:-"24h"}
    
    log_info "Cleaning up Docker resources (age: $age)"
    
    # Remove stopped containers
    log_debug "Removing stopped containers..."
    docker container prune -f --filter "until=$age"
    
    # Remove unused images
    log_debug "Removing unused images..."
    if [[ "$all" == "true" ]]; then
        docker image prune -af --filter "until=$age"
    else
        docker image prune -f --filter "until=$age"
    fi
    
    # Remove unused volumes
    log_debug "Removing unused volumes..."
    docker volume prune -f
    
    # Remove unused networks
    log_debug "Removing unused networks..."
    docker network prune -f --filter "until=$age"
    
    # Remove build cache
    if [[ "$all" == "true" ]]; then
        log_debug "Removing build cache..."
        docker builder prune -af --filter "until=$age"
    fi
    
    # Optional: Remove all dangling resources
    if [[ "$all" == "true" ]]; then
        log_debug "Removing dangling resources..."
        docker system prune -f --filter "until=$age"
    fi
    
    # Report disk space reclaimed
    if is_debug; then
        log_debug "Docker disk usage after cleanup:"
        docker system df
    fi
}
`

const gitLibTemplate = `#!/usr/bin/env bash
# Git utility functions for version management and repository operations
# Provides functions for semantic versioning, tagging, and repository status

source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

# Version pattern validation
readonly VERSION_PATTERN='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-((0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(\.(0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*))?(\+([0-9a-zA-Z-]+(\.[0-9a-zA-Z-]+)*))?$'

# Ensure we're in a git repository
ensure_git_repo() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "Not a git repository"
        exit 1
    fi
}

# Get the current version from git tags with fallback strategies
get_version() {
    local version

    ensure_git_repo
    
    if [[ -n "${CI_COMMIT_TAG:-}" ]]; then
        version="${CI_COMMIT_TAG}"
    elif [[ -n "${VERSION:-}" ]]; then
        version="${VERSION}"
    else
        version=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
        
        # Add branch name for non-release versions
        if [[ "$version" == "dev" ]]; then
            local branch
            branch=$(get_branch_name)
            version="dev-${branch}-$(get_short_commit_hash)"
        fi
    fi
    
    echo "$version"
}

# Get the latest tag with validation
get_latest_tag() {
    ensure_git_repo
    
    local latest_tag
    latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    
    # Validate tag format
    if ! echo "$latest_tag" | grep -qE "$VERSION_PATTERN"; then
        log_warn "Latest tag '$latest_tag' does not follow semantic versioning"
        latest_tag="v0.0.0"
    fi
    
    echo "$latest_tag"
}

# Calculate the next version based on semver rules
get_next_version() {
    local current_version
    local next_version
    local bump_type=${1:-patch}
    local pre_release=${2:-}
    
    current_version=$(get_latest_tag)
    current_version=${current_version#v}
    
    # Split version into components
    if ! IFS='.' read -r major minor patch prerelease <<< "${current_version/-/ }"; then
        log_error "Failed to parse version: $current_version"
        exit 1
    fi
    
    # Remove any build metadata
    patch=${patch%%+*}
    
    case "$bump_type" in
        major)
            next_version="$((major + 1)).0.0"
            ;;
        minor)
            next_version="${major}.$((minor + 1)).0"
            ;;
        patch)
            next_version="${major}.${minor}.$((patch + 1))"
            ;;
        *)
            log_error "Invalid bump type: $bump_type (expected: major|minor|patch)"
            exit 1
            ;;
    esac
    
    # Add pre-release suffix if specified
    if [[ -n "$pre_release" ]]; then
        next_version="${next_version}-${pre_release}"
    fi
    
    echo "v${next_version}"
}

# Create and push a new tag with validation
create_tag() {
    local version=$1
    local message=${2:-"Release ${version}"}
    local force=${3:-false}
    
    ensure_git_repo
    
    # Validate version format
    if ! echo "$version" | grep -qE "$VERSION_PATTERN"; then
        log_error "Invalid version format: $version"
        log_error "Version must match pattern: $VERSION_PATTERN"
        exit 1
    }
    
    # Check if tag already exists
    if git rev-parse "$version" >/dev/null 2>&1; then
        if [[ "$force" != "true" ]]; then
            log_error "Tag $version already exists. Use force=true to override"
            exit 1
        fi
        log_warn "Force-updating existing tag: $version"
        git tag -d "$version"
        git push origin ":refs/tags/$version"
    fi
    
    log_info "Creating tag: $version"
    if ! git tag -a "$version" -m "$message"; then
        log_error "Failed to create tag: $version"
        exit 1
    fi
    
    if ! git push origin "$version"; then
        log_error "Failed to push tag: $version"
        git tag -d "$version"
        exit 1
    fi
}

# Generate a changelog between tags
get_changelog() {
    local from_tag=${1:-$(get_latest_tag)}
    local to_ref=${2:-HEAD}
    local format=${3:-"* %s (%h)"}
    
    ensure_git_repo
    
    if ! git rev-parse "$from_tag" >/dev/null 2>&1; then
        log_error "Invalid from_tag: $from_tag"
        exit 1
    fi
    
    if ! git rev-parse "$to_ref" >/dev/null 2>&1; then
        log_error "Invalid to_ref: $to_ref"
        exit 1
    fi
    
    git log "${from_tag}..${to_ref}" --no-merges --pretty=format:"$format"
}

# Repository status functions
is_working_directory_clean() {
    ensure_git_repo
    [[ -z "$(git status --porcelain)" ]]
}

is_current_branch_main() {
    local branch
    branch=$(get_branch_name)
    [[ "$branch" == "main" || "$branch" == "master" ]]
}

has_uncommitted_changes() {
    ! is_working_directory_clean
}

# Repository information functions
get_branch_name() {
    ensure_git_repo
    git rev-parse --abbrev-ref HEAD
}

get_commit_hash() {
    ensure_git_repo
    git rev-parse HEAD
}

get_short_commit_hash() {
    ensure_git_repo
    git rev-parse --short HEAD
}

get_repo_root() {
    ensure_git_repo
    git rev-parse --show-toplevel
}

get_repo_name() {
    ensure_git_repo
    basename "$(get_repo_root)"
}

# Utility functions
sync_tags() {
    ensure_git_repo
    log_info "Syncing tags with remote"
    git fetch --tags --force
    git push --tags
}

cleanup_old_tags() {
    local keep_count=${1:-10}
    ensure_git_repo
    
    log_info "Cleaning up old tags (keeping last $keep_count)"
    local tags
    tags=$(git tag --sort=-creatordate | tail -n +$((keep_count + 1)))
    
    if [[ -n "$tags" ]]; then
        echo "$tags" | xargs -r git tag -d
        echo "$tags" | xargs -r git push origin --delete
    fi
}
`

const versionLibTemplate = `#!/usr/bin/env bash
# Version management functions

source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
source "$(dirname "${BASH_SOURCE[0]}")/git.sh"

# Get the build version
get_build_version() {
    if [[ -n "${VERSION:-}" ]]; then
        echo "${VERSION}"
    elif [[ -n "${CI_COMMIT_TAG:-}" ]]; then
        echo "${CI_COMMIT_TAG}"
    else
        echo "$(get_version)"
    fi
}

# Validate version format
validate_version() {
    local version=$1
    local version_regex="^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$"
    
    if [[ ! $version =~ $version_regex ]]; then
        log_error "Invalid version format: $version"
        log_error "Version must match semantic versioning format: vMAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]"
        return 1
    fi
}

# Compare two versions
compare_versions() {
    local version1=$1
    local version2=$2
    
    # Remove 'v' prefix if present
    version1=${version1#v}
    version2=${version2#v}
    
    if [[ "$version1" == "$version2" ]]; then
        echo "equal"
    elif [[ "$(printf '%s\n' "$version1" "$version2" | sort -V | head -n1)" == "$version1" ]]; then
            echo "less"
        else
            echo "greater"
        fi
}

# Get version components
get_version_components() {
    local version=$1
    version=${version#v}
    
    # Split version into components
    IFS='.-+' read -r major minor patch prerelease build <<< "$version"
    
    echo "MAJOR=$major"
    echo "MINOR=$minor"
    echo "PATCH=$patch"
    echo "PRERELEASE=$prerelease"
    echo "BUILD=$build"
}

# Generate version file
generate_version_file() {
    local version
    version=$(get_build_version)
    
    cat > "${PROJECT_ROOT}/version.go" << EOF
package main

var (
    version = "${version}"
    commit  = "$(get_commit_hash)"
    date    = "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
)
EOF
}
`
const buildTaskTemplate = `#!/usr/bin/env bash
# Build task script
# Builds the project binaries

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"
source "$(dirname "${BASH_SOURCE[0]}")/../lib/version.sh"

main() {
    local version
    version=$(get_build_version)
    
    log_info "Building version ${version}..."

    # Build flags
    local build_flags=(
        "-trimpath"
        "-ldflags=-s -w -X main.version=${version} -X main.commit=$(get_commit_hash) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    )

    # Add debug info in development
    if [[ "${GO_ENV:-}" == "development" ]]; then
        build_flags+=("-gcflags=all=-N -l")
    fi

    # Create bin directory
    mkdir -p "${PROJECT_ROOT}/bin"

    # Build for current platform
    build_binary "{{.ProjectName}}" "${build_flags[@]}"

    log_info "Build complete!"
}

build_binary() {
    local binary=$1
    shift
    local build_flags=("$@")

    log_info "Building ${binary}..."
    
    go build \
        "${build_flags[@]}" \
        -o "${PROJECT_ROOT}/bin/${binary}" \
        "${PROJECT_ROOT}/cmd/${binary}"
}

main "$@"
`

const testTaskTemplate = `#!/usr/bin/env bash
# Test task script
# Runs tests and generates coverage reports

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

main() {
    local coverage_dir="${PROJECT_ROOT}/coverage"
    mkdir -p "$coverage_dir"

    # Run tests with coverage
    run_tests "$coverage_dir"
    
    # Check coverage threshold
    check_coverage_threshold "$coverage_dir/coverage.out"
    
    # Generate coverage report
    generate_coverage_report "$coverage_dir"
}

run_tests() {
    local coverage_dir=$1
    
    log_info "Running tests..."
    
    go test \
        -race \
        -coverprofile="$coverage_dir/coverage.out" \
        -covermode=atomic \
        ./...
}

check_coverage_threshold() {
    local coverage_file=$1
    local threshold=${COVERAGE_THRESHOLD:-70}
    
    local coverage
    coverage=$(go tool cover -func="$coverage_file" | grep total: | awk '{print $3}' | sed 's/%//')
    
    log_info "Total coverage: ${coverage}%"
    
    if (( $(echo "$coverage < $threshold" | bc -l) )); then
        log_error "Coverage ${coverage}% is below threshold ${threshold}%"
        exit 1
    fi
}

generate_coverage_report() {
    local coverage_dir=$1
    
    log_info "Generating coverage report..."
    go tool cover -html="$coverage_dir/coverage.out" -o "$coverage_dir/coverage.html"
}

main "$@"
`

const lintTaskTemplate = `#!/usr/bin/env bash
# Lint task script
# Runs linters and code quality checks

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

main() {
    # Install golangci-lint if not present
    ensure_golangci_lint
    
    log_info "Running linters..."
    
    # Run golangci-lint
    golangci-lint run \
        --timeout=5m \
        --config="${PROJECT_ROOT}/.golangci.yml" \
        ./...
    
    # Run go fmt
    check_formatting
    
    # Run go vet
    go vet ./...
    
    log_info "Linting complete!"
}

ensure_golangci_lint() {
    if ! command -v golangci-lint >/dev/null 2>&1; then
        log_info "Installing golangci-lint..."
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    fi
}

check_formatting() {
    log_info "Checking code formatting..."
    
    local files
    files=$(gofmt -l .)
    
    if [[ -n "$files" ]]; then
        log_error "The following files are not properly formatted:"
        echo "$files"
        exit 1
    fi
}

main "$@"
`

const releaseTaskTemplate = `#!/usr/bin/env bash
# Release task script
# Handles version bumping and release creation

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"
source "$(dirname "${BASH_SOURCE[0]}")/../lib/git.sh"
source "$(dirname "${BASH_SOURCE[0]}")/../lib/version.sh"

main() {
    local bump_type=${1:-patch}
    local next_version
    
    # Ensure working directory is clean
    if ! is_working_directory_clean; then
        log_error "Working directory is not clean"
        exit 1
    }
    
    # Get next version
    next_version=$(get_next_version "$bump_type")
    
    # Validate version format
    if ! validate_version "$next_version"; then
        exit 1
    }
    
    # Generate changelog
    local changelog
    changelog=$(get_changelog)
    
    # Create and push tag
    create_tag "$next_version" "Release ${next_version}\n\n${changelog}"
    
    log_info "Release ${next_version} created successfully!"
    log_info "Changelog:\n${changelog}"
}

main "$@"
`

const dockerTaskTemplate = `#!/usr/bin/env bash
# Docker task script
# Handles Docker image building and publishing

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"
source "$(dirname "${BASH_SOURCE[0]}")/../lib/docker.sh"

main() {
    local cmd=${1:-build}
    local version
    version=$(get_build_version)
    
    case "$cmd" in
        build)
            build_images "$version"
            ;;
        push)
            push_images "$version"
            ;;
        *)
            log_error "Unknown command: $cmd"
            log_error "Usage: $0 [build|push]"
            exit 1
            ;;
    esac
}

build_images() {
    local version=$1
    local dockerfile="${PROJECT_ROOT}/build/docker/Dockerfile"
    local context="${PROJECT_ROOT}"
    
    # Build the main application image
    docker_build "{{.ProjectName}}:${version}" "$dockerfile" "$context"
    
    # Build any additional images (e.g., debug, minimal)
    if [[ -f "${PROJECT_ROOT}/build/docker/Dockerfile.debug" ]]; then
        docker_build "{{.ProjectName}}:${version}-debug" \
            "${PROJECT_ROOT}/build/docker/Dockerfile.debug" \
            "$context"
    fi
}

push_images() {
    local version=$1
    local registry=${DOCKER_REGISTRY:-""}
    
    # Push the main application image
    docker_push "{{.ProjectName}}:${version}" "$registry"
    
    # Push any additional images
    if [[ -f "${PROJECT_ROOT}/build/docker/Dockerfile.debug" ]]; then
        docker_push "{{.ProjectName}}:${version}-debug" "$registry"
    fi
}

main "$@"
`

const protoTaskTemplate = `#!/usr/bin/env bash
# Proto task script
# Generates code from protobuf definitions

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

generate_protos() {
    local proto_dir="${PROJECT_ROOT}/proto"
    local out_dir="${PROJECT_ROOT}/proto"
    
    log_info "Generating protobuf code..."
    
    # Clean existing generated files (only .pb.go files)
    find "${out_dir}" -type f -name "*.pb.go" -delete
    
    # Find all proto files
    local proto_files
    proto_files=$(find "${proto_dir}" -name "*.proto")
    
    # Generate Go code
    protoc \
        -I "${proto_dir}" \
        -I "/path/to/grpc-gateway"  // Replace with actual path
        -I "/path/to/protoc-gen-validate"  // Replace with actual path
        --go_out="${out_dir}" \
        --go_opt=paths=source_relative \
        --go-grpc_out="${out_dir}" \
        --go-grpc_opt=paths=source_relative \
        --grpc-gateway_out="${out_dir}" \
        --grpc-gateway_opt=paths=source_relative \
        --validate_out="lang=go,paths=source_relative:${out_dir}" \
        ${proto_files}
    
    # Clean up backup files
    find "${out_dir}" -type f -name "*.go.bak" -delete
}

install_proto_tools() {
    log_info "Installing protobuf tools"

    # Install protoc plugins
    go install \
        google.golang.org/protobuf/cmd/protoc-gen-go@latest \
        google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
        github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest \
        github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest \
        github.com/envoyproxy/protoc-gen-validate@latest

    # Install buf if available
    if command -v brew >/dev/null 2>&1; then
        brew install buf
    elif command -v go >/dev/null 2>&1; then
        go install github.com/bufbuild/buf/cmd/buf@latest
    fi
}

verify_proto_tools() {
    local missing_tools=()

    # Check required tools
    local tools=(
        "protoc"
        "protoc-gen-go"
        "protoc-gen-go-grpc"
        "protoc-gen-validate"
        "protoc-gen-grpc-gateway"
        "protoc-gen-openapiv2"
    )

    for tool in "${tools[@]}"; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            missing_tools+=("$tool")
        fi
    done

    if [ ${#missing_tools[@]} -gt 0 ]; then
        log_warn "Missing required tools: ${missing_tools[*]}"
        log_info "Installing missing tools..."
        install_proto_tools
    fi
}

main() {
    log_info "Starting protobuf generation"
    
    verify_proto_tools
    generate_protos
    
    log_info "Protobuf generation complete!"
}

main "$@"
`

const dependenciesTaskTemplate = `#!/usr/bin/env bash
# Dependencies task script
# Manages project dependencies and tools

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

update_dependencies() {
    log_info "Updating Go dependencies"
    go get -u ./...
    go mod tidy
}

verify_dependencies() {
    log_info "Verifying dependencies"
    go mod verify
}

install_tools() {
    log_info "Installing development tools..."
    
    # Run dependencies task with tools
    "${PROJECT_ROOT}/scripts/ci/tasks/dependencies.sh" --with-tools
    
    # Install additional development tools
    go install \
        github.com/cosmtrek/air@latest \
        github.com/go-delve/delve/cmd/dlv@latest \
        github.com/swaggo/swag/cmd/swag@latest
}

main() {
    log_info "Starting dependencies management"
    
    # Verify required tools
    ensure_command "go"
    
    update_dependencies
    verify_dependencies
    
    if [[ "${1:-}" == "--with-tools" ]]; then
        install_tools
    fi
    
    log_info "Dependencies management complete!"
}

main "$@"
`

const packageTaskTemplate = `#!/usr/bin/env bash
# Package task script
# Builds native packages (DEB, RPM, APK)

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

# Package metadata
PACKAGE_NAME="{{.ProjectName}}"
PACKAGE_VERSION="${VERSION:-0.0.1}"
PACKAGE_RELEASE="1"
PACKAGE_ARCH="amd64"
PACKAGE_DESCRIPTION="{{.ProjectName}} - {{.Description}}"
PACKAGE_MAINTAINER="{{.Author}}"
PACKAGE_LICENSE="{{.License}}"

# Directory structure
BUILD_DIR="${PROJECT_ROOT}/build"
PACKAGE_ROOT="${PROJECT_ROOT}/packaging"
STAGING_DIR="${PACKAGE_ROOT}/staging"
OUTPUT_DIR="${PROJECT_ROOT}/dist/packages"

create_package_dirs() {
    local dirs=(
        "${STAGING_DIR}/DEBIAN"
        "${STAGING_DIR}/usr/local/bin"
        "${STAGING_DIR}/etc/{{.ProjectName}}"
        "${STAGING_DIR}/usr/lib/systemd/system"
        "${STAGING_DIR}/etc/init.d"
        "${STAGING_DIR}/etc/conf.d"
        "${OUTPUT_DIR}"
    )

    for dir in "${dirs[@]}"; do
        mkdir -p "$dir"
    done
}

install_service_files() {
    local binary=$1
    local init_system=$2

    case "$init_system" in
        systemd)
            cp "${BUILD_DIR}/init/systemd/${binary}.service" \
                "${STAGING_DIR}/usr/lib/systemd/system/"
            ;;
        openrc)
            cp "${BUILD_DIR}/init/openrc/${binary}-openrc" \
                "${STAGING_DIR}/etc/init.d/${binary}"
            cp "${BUILD_DIR}/config/${binary}.conf" \
                "${STAGING_DIR}/etc/conf.d/${binary}"
            ;;
        sysvinit)
            cp "${BUILD_DIR}/init/sysvinit/${binary}-sysvinit" \
                "${STAGING_DIR}/etc/init.d/${binary}"
            ;;
    esac
}

install_config_files() {
    local binary=$1
    
    # Copy configuration files
    if [[ -d "${BUILD_DIR}/config" ]]; then
        cp -r "${BUILD_DIR}/config/"* "${STAGING_DIR}/etc/{{.ProjectName}}/"
    fi
}

build_tarball() {
    local binary=$1
    local format=${2:-}  # No default, empty if not provided
    log_info "Building tarball for ${binary}"

    local archive_name="${PACKAGE_NAME}-${PACKAGE_VERSION}-${PACKAGE_ARCH}"
    local temp_dir="${STAGING_DIR}/archive/${PACKAGE_NAME}-${PACKAGE_VERSION}"

    # Create directory structure
    mkdir -p "${temp_dir}"/{bin,etc,init/{systemd,openrc,sysvinit}}

    # Copy files
    cp "${PROJECT_ROOT}/bin/${binary}" "${temp_dir}/bin/"
    cp -r "${BUILD_DIR}/config/"* "${temp_dir}/etc/" 2>/dev/null || true
    cp -r "${BUILD_DIR}/init/"* "${temp_dir}/init/" 2>/dev/null || true
    cp "${PROJECT_ROOT}/README.md" "${temp_dir}/" 2>/dev/null || true
    cp "${PROJECT_ROOT}/LICENSE" "${temp_dir}/" 2>/dev/null || true

    # If no format specified, create tar.gz
    if [[ -z "$format" ]]; then
        tar -czf "${OUTPUT_DIR}/${archive_name}.tar.gz" \
            -C "${STAGING_DIR}/archive" \
            "${PACKAGE_NAME}-${PACKAGE_VERSION}"
        return
    fi

    # Handle specific format if provided
    case "$format" in
        gz|gzip)
            tar -czf "${OUTPUT_DIR}/${archive_name}.tar.gz" \
                -C "${STAGING_DIR}/archive" \
                "${PACKAGE_NAME}-${PACKAGE_VERSION}"
            ;;
        bz2|bzip2)
            tar -cjf "${OUTPUT_DIR}/${archive_name}.tar.bz2" \
                -C "${STAGING_DIR}/archive" \
                "${PACKAGE_NAME}-${PACKAGE_VERSION}"
            ;;
        xz)
            tar -cJf "${OUTPUT_DIR}/${archive_name}.tar.xz" \
                -C "${STAGING_DIR}/archive" \
                "${PACKAGE_NAME}-${PACKAGE_VERSION}"
            ;;
        zst|zstd)
            tar -cf - -C "${STAGING_DIR}/archive" "${PACKAGE_NAME}-${PACKAGE_VERSION}" | \
                zstd -T0 > "${OUTPUT_DIR}/${archive_name}.tar.zst"
            ;;
        zip)
            (cd "${STAGING_DIR}/archive" && \
                zip -r "${OUTPUT_DIR}/${archive_name}.zip" "${PACKAGE_NAME}-${PACKAGE_VERSION}")
            ;;
        *)
            log_error "Unsupported compression format: $format"
            return 1
            ;;
    esac
}

main() {
    log_info "Starting package generation"

    # Clean and create directories
    rm -rf "${STAGING_DIR}" "${OUTPUT_DIR}"
    create_package_dirs

    # Verify required tools
    if command -v dpkg-deb >/dev/null; then
        build_deb "{{.ProjectName}}"
    else
        log_warn "dpkg-deb not found, skipping DEB package"
    fi

    if command -v rpmbuild >/dev/null; then
        build_rpm "{{.ProjectName}}"
    else
        log_warn "rpmbuild not found, skipping RPM package"
    fi

    if command -v abuild >/dev/null; then
        build_apk "{{.ProjectName}}"
    else
        log_warn "abuild not found, skipping APK package"
    fi

    # Always build tarball as fallback
    build_tarball "{{.ProjectName}}"

    log_info "Package generation complete! Packages available in ${OUTPUT_DIR}"
}

main "$@"
`

const healthCheckTemplate = `#!/usr/bin/env bash
# Health check script
# Verifies the application is running correctly

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

check_process() {
    local binary=$1
    local pid_file="/var/run/${binary}.pid"

    if [[ -f "$pid_file" ]]; then
        local pid
        pid=$(cat "$pid_file")
        if kill -0 "$pid" 2>/dev/null; then
            log_info "Process is running (PID: $pid)"
            return 0
        fi
    fi

    log_error "Process is not running"
        return 1
}

check_http_endpoint() {
    local url=$1
    local expected_status=${2:-200}
    
    local response
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url")
    
    if [[ "$response" == "$expected_status" ]]; then
        log_info "HTTP endpoint is healthy"
        return 0
    fi
    
    log_error "HTTP endpoint returned status $response (expected $expected_status)"
        return 1
}

main() {
    local binary="{{.ProjectName}}"
    local health_url="http://localhost:8080/health"
    
    # Check process
    check_process "$binary"
    
    # Check HTTP endpoint
    check_http_endpoint "$health_url"
}

main "$@"
`

const cleanupTemplate = `#!/usr/bin/env bash
# Cleanup script
# Removes temporary files and build artifacts

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

# Cleanup functions for different artifact types
cleanup_build_artifacts() {
    local dry_run=$1
    local paths=(
        "${PROJECT_ROOT}/bin"
        "${PROJECT_ROOT}/dist"
        "${PROJECT_ROOT}/.build-cache"
    )
    
    log_info "Cleaning build artifacts..."
    for path in "${paths[@]}"; do
        if [[ "$dry_run" == "true" ]]; then
            log_info "[DRY RUN] Would remove: $path"
        else
            rm -rf "$path" && log_debug "Removed: $path"
        fi
    done
}

cleanup_test_artifacts() {
    local dry_run=$1
    local paths=(
        "${PROJECT_ROOT}/coverage"
        "${PROJECT_ROOT}/.test-cache"
    )
    
    log_info "Cleaning test artifacts..."
    # Remove test coverage directories
    for path in "${paths[@]}"; do
        if [[ "$dry_run" == "true" ]]; then
            log_info "[DRY RUN] Would remove: $path"
        else
            rm -rf "$path" && log_debug "Removed: $path"
        fi
    done
    
    # Find and remove test binaries
    if [[ "$dry_run" == "true" ]]; then
        log_info "[DRY RUN] Would remove test binaries:"
        find "${PROJECT_ROOT}" -type f -name "*.test" -print
    else
        find "${PROJECT_ROOT}" -type f -name "*.test" -delete -print | while read -r file; do
            log_debug "Removed: $file"
        done
    fi
}

cleanup_generated_files() {
    local dry_run=$1
    local paths=(
        "${PROJECT_ROOT}/pkg/gen"
        "${PROJECT_ROOT}/api/gen"
        "${PROJECT_ROOT}/internal/gen"
        "${PROJECT_ROOT}/docs/gen"
    )
    
    log_info "Cleaning generated files..."
    for path in "${paths[@]}"; do
        if [[ "$dry_run" == "true" ]]; then
            log_info "[DRY RUN] Would remove: $path"
        else
            rm -rf "$path" && log_debug "Removed: $path"
        fi
    done
}

cleanup_dependencies() {
    local dry_run=$1
    
    log_info "Cleaning dependencies..."
    if [[ "$dry_run" == "true" ]]; then
        log_info "[DRY RUN] Would clean Go mod cache"
        log_info "[DRY RUN] Would remove vendor directory"
    else
        go clean -modcache && log_debug "Cleaned Go mod cache"
        rm -rf "${PROJECT_ROOT}/vendor" && log_debug "Removed vendor directory"
    fi
}

cleanup_docker() {
    local dry_run=$1
    
    log_info "Cleaning Docker artifacts..."
    if [[ "$dry_run" == "true" ]]; then
        log_info "[DRY RUN] Would clean Docker build cache"
    else
        if command -v docker >/dev/null 2>&1; then
            docker system prune -f --filter "label=project={{.ProjectName}}" && log_debug "Cleaned Docker build cache"
        else
            log_warn "Docker not found, skipping Docker cleanup"
        fi
    fi
}

cleanup_ide() {
    local dry_run=$1
    local paths=(
        "${PROJECT_ROOT}/.idea"
        "${PROJECT_ROOT}/.vscode"
        "${PROJECT_ROOT}/.vs"
        "${PROJECT_ROOT}/*.iml"
        "${PROJECT_ROOT}/.settings"
        "${PROJECT_ROOT}/.project"
        "${PROJECT_ROOT}/.classpath"
    )
    
    log_info "Cleaning IDE files..."
    for path in "${paths[@]}"; do
        if [[ "$dry_run" == "true" ]]; then
            log_info "[DRY RUN] Would remove: $path"
        else
            rm -rf "$path" && log_debug "Removed: $path"
        fi
    done
}

print_disk_usage() {
    local before=$1
    local after=$2
    local saved
    saved=$((before - after))
    
    log_info "Disk usage summary:"
    log_info "  Before: $(numfmt --to=iec-i --suffix=B $before)"
    log_info "  After:  $(numfmt --to=iec-i --suffix=B $after)"
    log_info "  Saved:  $(numfmt --to=iec-i --suffix=B $saved)"
}

main() {
    local clean_all=false
    local dry_run=false
    local clean_docker=false
    local clean_ide=false
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --all)
                clean_all=true
                shift
                ;;
            --dry-run)
                dry_run=true
                shift
                ;;
            --docker)
                clean_docker=true
                shift
                ;;
            --ide)
                clean_ide=true
                shift
                ;;
            -h|--help)
                echo "Usage: $0 [options]"
                echo "Options:"
                echo "  --all      Clean everything, including dependencies"
                echo "  --dry-run  Show what would be cleaned without actually removing"
                echo "  --docker   Clean Docker build cache"
                echo "  --ide      Clean IDE-specific files"
                echo "  --help     Show this help message"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Calculate initial disk usage
    local disk_usage_before
    disk_usage_before=$(du -sb "${PROJECT_ROOT}" 2>/dev/null | cut -f1)
    
    # Run cleanup functions
    cleanup_build_artifacts "$dry_run"
    cleanup_test_artifacts "$dry_run"
    cleanup_generated_files "$dry_run"
    
    if [[ "$clean_all" == "true" ]]; then
        cleanup_dependencies "$dry_run"
    fi
    
    if [[ "$clean_docker" == "true" ]]; then
        cleanup_docker "$dry_run"
    fi
    
    if [[ "$clean_ide" == "true" ]]; then
        cleanup_ide "$dry_run"
    fi
    
    # Calculate final disk usage and print summary
    if [[ "$dry_run" != "true" ]]; then
        local disk_usage_after
        disk_usage_after=$(du -sb "${PROJECT_ROOT}" 2>/dev/null | cut -f1)
        print_disk_usage "$disk_usage_before" "$disk_usage_after"
    fi
    
    log_info "Cleanup ${dry_run:+[DRY RUN] }complete!"
}

main "$@"
`

const setupDevTemplate = `#!/usr/bin/env bash
# Setup development environment
# Installs required tools and dependencies

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

setup_git_hooks() {
    log_info "Setting up Git hooks..."
    
    local hooks_dir="${PROJECT_ROOT}/.git/hooks"
    local ci_hooks_dir="${PROJECT_ROOT}/scripts/ci/hooks"
    
    # Create hooks directory if it doesn't exist
    mkdir -p "$hooks_dir"
    
    # Link all hooks
    for hook in "$ci_hooks_dir"/*; do
        if [[ -f "$hook" ]]; then
            local hook_name
            hook_name=$(basename "$hook")
            ln -sf "$hook" "${hooks_dir}/${hook_name}"
            chmod +x "${hooks_dir}/${hook_name}"
        fi
    done
}

setup_tools() {
    log_info "Installing development tools..."
    
    # Run dependencies task with tools
    "${PROJECT_ROOT}/scripts/ci/tasks/dependencies.sh" --with-tools
    
    # Install additional development tools
    go install \
        github.com/cosmtrek/air@latest \
        github.com/go-delve/delve/cmd/dlv@latest \
        github.com/swaggo/swag/cmd/swag@latest
}

setup_env() {
    log_info "Setting up environment..."
    
    local env_file="${PROJECT_ROOT}/.env"
    local example_env="${PROJECT_ROOT}/.env.example"
    local template_env="${PROJECT_ROOT}/.env.template"
    
    # Check for existing .env file
    if [[ -f "$env_file" ]]; then
        log_info "Existing .env file found"
        
        # Optional: Check for missing variables
        if [[ -f "$example_env" ]]; then
            local missing_vars
            missing_vars=$(grep -v '^#' "$example_env" | cut -d '=' -f1 | while read -r var; do
                grep -q "^${var}=" "$env_file" || echo "$var"
            done)
            
            if [[ -n "$missing_vars" ]]; then
                log_warn "Missing environment variables in .env file:"
                echo "$missing_vars" | sed 's/^/  - /'
            fi
        fi
        
        return 0
    fi
    
    # Try to create .env file from available templates
    if [[ -f "$example_env" ]]; then
        log_info "Creating .env file from .env.example"
        cp "$example_env" "$env_file"
    elif [[ -f "$template_env" ]]; then
        log_info "Creating .env file from .env.template"
        cp "$template_env" "$env_file"
    else
        log_warn "No environment template found (.env.example or .env.template)"
        log_info "Creating minimal .env file"
        cat > "$env_file" << EOF
# Environment Configuration
# Generated on $(date -u +"%Y-%m-%d %H:%M:%S UTC")

# Application
APP_ENV=development
APP_DEBUG=true
APP_PORT=8080

# Add your environment variables below
EOF
    fi
    
    # Set appropriate permissions
    chmod 600 "$env_file"
    
    # Validate environment file
    if ! grep -q "APP_ENV=" "$env_file"; then
        log_warn "Environment file may be missing critical variables"
    fi
    
    log_info "Environment file created at: $env_file"
    log_info "Please review and update the environment variables as needed"
}

main() {
    log_info "Setting up development environment..."
    
    setup_git_hooks
    setup_tools
    setup_env
    
    log_info "Development environment setup complete!"
}

main "$@"
`

const mainBuildTemplate = `#!/usr/bin/env bash
# Main build script
# Entry point for building the project
# Provides a unified interface for all build-related operations

source "$(dirname "${BASH_SOURCE[0]}")/lib/common.sh"
source "$(dirname "${BASH_SOURCE[0]}")/lib/version.sh"

# Print usage information
usage() {
    cat << EOF
Usage: $0 <command> [options]

Commands:
    build       Build the project binaries
    docker      Build Docker images
    package     Create distribution packages
    proto       Generate protobuf code
    all         Run all build steps
    clean       Clean build artifacts
    help        Show this help message

Options:
    -v, --verbose     Enable verbose output
    -d, --debug       Build with debug information
    -r, --release     Build for release
    -p, --platform    Specify build platform (e.g., linux/amd64)
    -o, --output      Specify output directory
    --version         Show version information

Examples:
    $0 build --debug
    $0 docker --platform linux/amd64,linux/arm64
    $0 package --release
EOF
    exit 1
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -v|--verbose)
                export LOG_LEVEL=DEBUG
                shift
                ;;
            -d|--debug)
                export DEBUG=true
                shift
                ;;
            -r|--release)
                export RELEASE=true
                shift
                ;;
            -p|--platform)
                export BUILD_PLATFORM="$2"
                shift 2
                ;;
            -o|--output)
                export OUTPUT_DIR="$2"
                shift 2
                ;;
            --version)
                echo "Version: $(get_build_version)"
                exit 0
                ;;
            -h|--help)
                usage
                ;;
            *)
                COMMAND="$1"
                shift
                ARGS=("$@")
                break
                ;;
        esac
    done
}

# Run all build steps
build_all() {
    local steps=(
        "proto"
        "build"
        "docker"
        "package"
    )

    for step in "${steps[@]}"; do
        log_info "Running build step: $step"
        if ! run_build_step "$step"; then
            log_error "Build step '$step' failed"
            return 1
        fi
    done
}

# Run a single build step
run_build_step() {
    local step=$1
    local script="${PROJECT_ROOT}/scripts/ci/tasks/${step}.sh"

    if [[ ! -f "$script" ]]; then
        log_error "Build script not found: $script"
        return 1
    fi

    if [[ -n "${DEBUG:-}" ]]; then
        log_debug "Running build step '$step' with args: ${ARGS[*]:-}"
    fi

    "$script" "${ARGS[@]:-}"
}

# Clean build artifacts
clean_build() {
    log_info "Cleaning build artifacts"
    "${PROJECT_ROOT}/scripts/ci/utils/cleanup.sh" --build
}

main() {
    # Ensure we're in the project root
    cd "${PROJECT_ROOT}" || exit 1

    # Parse command line arguments
    parse_args "$@"

    # Set default command if none provided
    COMMAND=${COMMAND:-build}

    # Handle commands
    case "$COMMAND" in
        build|docker|package|proto)
            run_build_step "$COMMAND"
            ;;
        all)
            build_all
            ;;
        clean)
            clean_build
            ;;
        help)
            usage
            ;;
        *)
            log_error "Unknown command: $COMMAND"
            usage
            ;;
    esac
}

# Ensure proper error handling
set -euo pipefail
trap 'log_error "Error occurred in build script"' ERR

main "$@"
`

const mainTestTemplate = `#!/usr/bin/env bash
# Main test script
# Entry point for testing the project

source "$(dirname "${BASH_SOURCE[0]}")/lib/common.sh"

main() {
    local cmd=${1:-test}
    shift || true
    
    case "$cmd" in
        test)
            "${PROJECT_ROOT}/scripts/ci/tasks/test.sh" "$@"
            ;;
        lint)
            "${PROJECT_ROOT}/scripts/ci/tasks/lint.sh" "$@"
            ;;
        coverage)
            "${PROJECT_ROOT}/scripts/ci/tasks/test.sh" --coverage "$@"
            ;;
        all)
            log_info "Running all test tasks..."
            
            # Run linting first to catch any style issues
            if ! "${PROJECT_ROOT}/scripts/ci/tasks/lint.sh" "$@"; then
                log_error "Linting failed"
                exit 1
            fi
            
            # Run tests with coverage
            if ! "${PROJECT_ROOT}/scripts/ci/tasks/test.sh" --coverage "$@"; then
                log_error "Tests failed"
                exit 1
            fi
            
            log_info "All test tasks completed successfully!"
            ;;
        *)
            log_error "Unknown command: $cmd"
            log_error "Usage: $0 [test|lint|coverage|all]"
            exit 1
            ;;
    esac
}

main "$@"
`

const mainCITemplate = `#!/usr/bin/env bash
# Main CI script
# Entry point for CI pipeline

source "$(dirname "${BASH_SOURCE[0]}")/lib/common.sh"

main() {
    log_info "Starting CI pipeline"
    
    # Install dependencies and tools
    "${PROJECT_ROOT}/scripts/ci/tasks/dependencies.sh" --with-tools
    
    # Run linters
    "${PROJECT_ROOT}/scripts/ci/tasks/lint.sh"
    
    # Run tests with coverage
    "${PROJECT_ROOT}/scripts/ci/tasks/test.sh" --coverage
    
    # Build binaries
    "${PROJECT_ROOT}/scripts/ci/tasks/build.sh"
    
    # Build Docker images if needed
    if [[ -f "${PROJECT_ROOT}/build/docker/Dockerfile" ]]; then
        "${PROJECT_ROOT}/scripts/ci/tasks/docker.sh" build
        
        # Push images if on main branch or tag
        if [[ "${CI_COMMIT_BRANCH}" == "main" ]] || [[ -n "${CI_COMMIT_TAG}" ]]; then
            "${PROJECT_ROOT}/scripts/ci/tasks/docker.sh" push
        fi
    fi
    
    # Create packages if this is a release
    if [[ -n "${CI_COMMIT_TAG}" ]]; then
        "${PROJECT_ROOT}/scripts/ci/tasks/package.sh"
    fi
    
    log_info "CI pipeline completed successfully!"
}

main "$@"
`

const makefileTemplate2 = `# Makefile for CI Scripts

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
`

const taskfileTemplate = `# Taskfile.yml
version: '3'

vars:
  CI_SCRIPTS: ./scripts/ci

env:
  PROJECT_ROOT: '{{ "{{" }}.ROOT{{ "}}" }}'

tasks:
  default:
    cmds:
      - task: test
      - task: build

  test:
    desc: Run tests
    cmds:
      - '{{ "{{" }}.CI_SCRIPTS{{ "}}" }}/test {{ "{{" }}.CLI_ARGS{{ "}}" }}'
    sources:
      - 'pkg/**/*.go'
      - 'cmd/**/*.go'
    generates:
      - coverage/coverage.out

  test:watch:
    desc: Run tests in watch mode
    cmds:
      - '{{ "{{" }}.CI_SCRIPTS{{ "}}" }}/test --watch {{ "{{" }}.CLI_ARGS{{ "}}" }}'

  test:coverage:
    desc: Run tests with coverage
    cmds:
      - '{{ "{{" }}.CI_SCRIPTS{{ "}}" }}/test coverage {{ "{{" }}.CLI_ARGS{{ "}}" }}'

  build:
    desc: Build project
    cmds:
      - '{{ "{{" }}.CI_SCRIPTS{{ "}}" }}/build {{ "{{" }}.CLI_ARGS{{ "}}" }}'
    sources:
      - 'pkg/**/*.go'
      - 'cmd/**/*.go'
    generates:
      - bin/{{ "{{" }}.PROJECT_NAME{{ "}}" }}

  build:debug:
    desc: Build with debug info
    env:
      DEBUG: "1"
    cmds:
      - task: build

  docker:
    desc: Build Docker images
    cmds:
      - '{{ "{{" }}.CI_SCRIPTS{{ "}}" }}/build docker {{ "{{" }}.CLI_ARGS{{ "}}" }}'

  package:
    desc: Create distribution packages
    deps: [build]
    cmds:
      - '{{ "{{" }}.CI_SCRIPTS{{ "}}" }}/build package {{ "{{" }}.CLI_ARGS{{ "}}" }}'

  ci:
    desc: Run full CI pipeline
    cmds:
      - '{{ "{{" }}.CI_SCRIPTS{{ "}}" }}/ci {{ "{{" }}.CLI_ARGS{{ "}}" }}'

  setup-dev:
    desc: Setup development environment
    cmds:
      - '{{ "{{" }}.CI_SCRIPTS{{ "}}" }}/setup-dev.sh'

  lint:
    desc: Run linters
    cmds:
      - '{{ "{{" }}.CI_SCRIPTS{{ "}}" }}/test lint {{ "{{" }}.CLI_ARGS{{ "}}" }}'

  generate:
    desc: Generate code
    cmds:
      - '{{ "{{" }}.CI_SCRIPTS{{ "}}" }}/tasks/generate.sh {{ "{{" }}.CLI_ARGS{{ "}}" }}'

  ci:test:
    desc: Test CI configurations
    cmds:
      - '{{ "{{" }}.CI_SCRIPTS{{ "}}" }}/utils/ci-tester.sh {{ "{{" }}.CLI_ARGS{{ "}}" }}'

  ci:test:github:
    desc: Test GitHub Actions
    cmds:
      - '{{ "{{" }}.CI_SCRIPTS{{ "}}" }}/utils/ci-tester.sh github {{ "{{" }}.CLI_ARGS{{ "}}" }}'

  ci:test:gitlab:
    desc: Test GitLab CI
    cmds:
      - '{{ "{{" }}.CI_SCRIPTS{{ "}}" }}/utils/ci-tester.sh gitlab {{ "{{" }}.CLI_ARGS{{ "}}" }}'

  db:
    desc: Database operations
    cmds:
      - '{{ "{{" }}.CI_SCRIPTS{{ "}}" }}/utils/db.sh {{ "{{" }}.CLI_ARGS{{ "}}" }}'

  release:
    desc: Create a release
    cmds:
      - '{{ "{{" }}.CI_SCRIPTS{{ "}}" }}/tasks/release.sh {{ "{{" }}.CLI_ARGS{{ "}}" }}'

  clean:
    desc: Clean build artifacts
    cmds:
      - '{{ "{{" }}.CI_SCRIPTS{{ "}}" }}/utils/cleanup.sh {{ "{{" }}.CLI_ARGS{{ "}}" }}'
`

const justfileTemplate = `# justfile

# Default recipe
default: test build

# Set environment variables
export PROJECT_ROOT := env_var_or_default("PROJECT_ROOT", justfile_directory())
export CI_SCRIPTS := PROJECT_ROOT + "/scripts/ci"

# Show available recipes
help:
    @just --list

# Test recipes
test *args:
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/test {{ "{{" }}args{{ "}}" }}

test-watch *args:
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/test --watch {{ "{{" }}args{{ "}}" }}

test-coverage *args:
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/test coverage {{ "{{" }}args{{ "}}" }}

# Build recipes
build *args:
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/build {{ "{{" }}args{{ "}}" }}

build-debug *args:
    DEBUG=1 {{ "{{" }}CI_SCRIPTS{{ "}}" }}/build {{ "{{" }}args{{ "}}" }}

docker *args:
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/build docker {{ "{{" }}args{{ "}}" }}

package *args:
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/build package {{ "{{" }}args{{ "}}" }}

# CI recipes
ci *args:
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/ci {{ "{{" }}args{{ "}}" }}

ci-test *args:
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/utils/ci-tester.sh {{ "{{" }}args{{ "}}" }}

ci-test-github *args:
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/utils/ci-tester.sh github {{ "{{" }}args{{ "}}" }}

ci-test-gitlab *args:
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/utils/ci-tester.sh gitlab {{ "{{" }}args{{ "}}" }}

# Development recipes
setup-dev:
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/setup-dev.sh

lint *args:
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/test lint {{ "{{" }}args{{ "}}" }}

generate *args:
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/tasks/generate.sh {{ "{{" }}args{{ "}}" }}

# Database recipes
db-start *args:
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/utils/db.sh start {{ "{{" }}args{{ "}}" }}

db-migrate *args:
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/utils/db.sh migrate {{ "{{" }}args{{ "}}" }}

db-seed *args:
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/utils/db.sh seed {{ "{{" }}args{{ "}}" }}

# Release recipes
release *args:
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/tasks/release.sh {{ "{{" }}args{{ "}}" }}

release-rc *args:
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/tasks/release.sh --rc {{ "{{" }}args{{ "}}" }}

release-hotfix *args:
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/tasks/release.sh --hotfix {{ "{{" }}args{{ "}}" }}

# Cleanup recipe
clean *args:
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/utils/cleanup.sh {{ "{{" }}args{{ "}}" }}

# Composite recipes
full-test: lint test test-coverage

full-build: generate build docker package

full-release: full-test full-build release

# Development workflow recipes
dev: setup-dev
    {{ "{{" }}CI_SCRIPTS{{ "}}" }}/utils/dev-server.sh

# CI testing workflow
test-all-ci: ci-test-github ci-test-gitlab
    @echo "All CI configurations tested successfully"

# Database workflow
db-reset: db-start db-migrate db-seed
    @echo "Database reset complete"
`

const magefileTemplate = `//go:build mage
package main

import (
    "github.com/magefile/mage/mg"
    "github.com/magefile/mage/sh"
    "os"
    "path/filepath"
)

const ciScripts = "./scripts/ci"

// Default target to run when none is specified
var Default = Build

// Test runs the test suite
func Test() error {
    return sh.RunV(filepath.Join(ciScripts, "test"))
}

// TestWatch runs tests in watch mode
func TestWatch() error {
    return sh.RunV(filepath.Join(ciScripts, "test"), "--watch")
}

// Build builds the project
func Build() error {
    return sh.RunV(filepath.Join(ciScripts, "build"))
}

// Docker builds Docker images
func Docker() error {
    return sh.RunV(filepath.Join(ciScripts, "build"), "docker")
}

// CI runs the full CI pipeline
func CI() error {
    return sh.RunV(filepath.Join(ciScripts, "ci"))
}

// CITest tests CI configurations locally
func CITest() error {
    return sh.RunV(filepath.Join(ciScripts, "utils", "ci-tester.sh"))
}

// Clean removes build artifacts
func Clean() error {
    return sh.RunV(filepath.Join(ciScripts, "utils", "cleanup.sh"))
}

// Release creates a new release
func Release() error {
    return sh.RunV(filepath.Join(ciScripts, "tasks", "release.sh"))
}

// Namespace example
type DB mg.Namespace

// Start starts the database
func (DB) Start() error {
    return sh.RunV(filepath.Join(ciScripts, "utils", "db.sh"), "start")
}

// Migrate runs database migrations
func (DB) Migrate() error {
    return sh.RunV(filepath.Join(ciScripts, "utils", "db.sh"), "migrate")
}

// Seed seeds the database
func (DB) Seed() error {
    return sh.RunV(filepath.Join(ciScripts, "utils", "db.sh"), "seed")
}
`

const rakefileTemplate = `# Rakefile
require 'rake'

CI_SCRIPTS = './scripts/ci'

# Default task
task default: %w[test build]

# Test tasks
desc 'Run tests'
task :test, [:args] do |t, args|
  sh "#{CI_SCRIPTS}/test #{args[:args]}"
end

desc 'Run tests in watch mode'
task :test_watch do
  sh "#{CI_SCRIPTS}/test --watch"
end

desc 'Run tests with coverage'
task :test_coverage do
  sh "#{CI_SCRIPTS}/test coverage"
end

# Build tasks
desc 'Build project'
task :build, [:args] do |t, args|
  sh "#{CI_SCRIPTS}/build #{args[:args]}"
end

desc 'Build Docker images'
task :docker do
  sh "#{CI_SCRIPTS}/build docker"
end

# CI tasks
desc 'Run CI pipeline'
task :ci do
  sh "#{CI_SCRIPTS}/ci"
end

desc 'Test CI configurations'
task :ci_test do
  sh "#{CI_SCRIPTS}/utils/ci-tester.sh"
end

namespace :ci do
  desc 'Test GitHub Actions'
  task :test_github do
    sh "#{CI_SCRIPTS}/utils/ci-tester.sh github"
  end

  desc 'Test GitLab CI'
  task :test_gitlab do
    sh "#{CI_SCRIPTS}/utils/ci-tester.sh gitlab"
  end
end

# Database tasks
namespace :db do
  desc 'Start database'
  task :start do
    sh "#{CI_SCRIPTS}/utils/db.sh start"
  end

  desc 'Run migrations'
  task :migrate do
    sh "#{CI_SCRIPTS}/utils/db.sh migrate"
  end

  desc 'Seed database'
  task :seed do
    sh "#{CI_SCRIPTS}/utils/db.sh seed"
  end

  desc 'Reset database'
  task reset: [:start, :migrate, :seed]
end

# Release tasks
desc 'Create release'
task :release, [:type] do |t, args|
  type_arg = args[:type] ? "--#{args[:type]}" : ""
  sh "#{CI_SCRIPTS}/tasks/release.sh #{type_arg}"
end
`

const invokeTemplate = `# tasks.py
from invoke import task, Collection

CI_SCRIPTS = './scripts/ci'

# Test tasks
@task(help={'args': 'Additional arguments for test command'})
def test(ctx, args=''):
    """Run tests"""
    ctx.run(f"{CI_SCRIPTS}/test {args}")

@task
def test_watch(ctx):
    """Run tests in watch mode"""
    ctx.run(f"{CI_SCRIPTS}/test --watch")

@task
def test_coverage(ctx):
    """Run tests with coverage"""
    ctx.run(f"{CI_SCRIPTS}/test coverage")

# Build tasks
@task(help={'args': 'Additional arguments for build command'})
def build(ctx, args=''):
    """Build project"""
    ctx.run(f"{CI_SCRIPTS}/build {args}")

@task
def docker(ctx):
    """Build Docker images"""
    ctx.run(f"{CI_SCRIPTS}/build docker")

# CI tasks
@task
def ci(ctx):
    """Run CI pipeline"""
    ctx.run(f"{CI_SCRIPTS}/ci")

@task
def ci_test(ctx, platform=''):
    """Test CI configurations"""
    cmd = f"{CI_SCRIPTS}/utils/ci-tester.sh"
    if platform:
        cmd += f" {platform}"
    ctx.run(cmd)

# Database tasks
@task
def db_start(ctx):
    """Start database"""
    ctx.run(f"{CI_SCRIPTS}/utils/db.sh start")

@task
def db_migrate(ctx):
    """Run database migrations"""
    ctx.run(f"{CI_SCRIPTS}/utils/db.sh migrate")

@task
def db_seed(ctx):
    """Seed database"""
    ctx.run(f"{CI_SCRIPTS}/utils/db.sh seed")

@task(db_start, db_migrate, db_seed)
def db_reset(ctx):
    """Reset database"""
    print("Database reset complete")

# Release tasks
@task(help={'type': 'Release type (rc, hotfix)'})
def release(ctx, type=''):
    """Create release"""
    type_arg = f"--{type}" if type else ""
    ctx.run(f"{CI_SCRIPTS}/tasks/release.sh {type_arg}")

# Create namespaces
ns = Collection()
ns.add_task(test)
ns.add_task(test_watch)
ns.add_task(test_coverage)
ns.add_task(build)
ns.add_task(docker)
ns.add_task(ci)
ns.add_task(ci_test)

# Database namespace
db = Collection('db')
db.add_task(db_start, 'start')
db.add_task(db_migrate, 'migrate')
db.add_task(db_seed, 'seed')
db.add_task(db_reset, 'reset')
ns.add_collection(db)
`

const daggerTemplate = `package main

import (
    "dagger.io/dagger"
    "universe.dagger.io/docker"
    "universe.dagger.io/bash"
)

// Project configuration
#Config: {
    projectName: string | *"{{.ProjectName}}"
}

config: #Config

// Base container with CI scripts
#BaseContainer: {
    docker.#Build & {
        steps: [
            docker.#Pull & {
                source: "golang:1.21"
            },
            docker.#Copy & {
                contents: client.filesystem."./".read.contents
                dest: "/workspace"
            },
            // Ensure scripts are executable
            bash.#Run & {
                script: contents: """
                    chmod -R +x /workspace/scripts/ci
                    """
            },
        ]
    }
}

dagger.#Plan & {
    client: {
        filesystem: {
            "./": read: {
                contents: dagger.#FS
                exclude: [
                    "bin",
                    "dist",
                    "coverage",
                    ".git",
                ]
            }
        }
        env: {
            CI:               string | *"true"
            CI_COMMIT_SHA:    string | *""
            CI_COMMIT_TAG:    string | *""
            CI_COMMIT_BRANCH: string | *""
            LOG_LEVEL:        string | *"INFO"
        }
    }

    actions: {
        // Build task using build.sh
        build: docker.#Run & {
            input: #BaseContainer
            workdir: "/workspace"
            command: {
                name: "./scripts/ci/tasks/build.sh"
            }
        }

        // Test task using test.sh
        test: docker.#Run & {
            input: #BaseContainer
            workdir: "/workspace"
            command: {
                name: "./scripts/ci/tasks/test.sh"
            }
        }

        // Lint task using lint.sh
        lint: docker.#Run & {
            input: #BaseContainer
            workdir: "/workspace"
            command: {
                name: "./scripts/ci/tasks/lint.sh"
            }
        }

        // Docker task using docker.sh
        docker: {
            build: docker.#Run & {
                input: #BaseContainer
                workdir: "/workspace"
                command: {
                    name: "./scripts/ci/tasks/docker.sh"
                    args: ["build"]
                }
            }

            push: docker.#Run & {
                input: #BaseContainer
                workdir: "/workspace"
                command: {
                    name: "./scripts/ci/tasks/docker.sh"
                    args: ["push"]
                }
            }
        }
        }

        // Release task using release.sh
        release: docker.#Run & {
            input: #BaseContainer
            workdir: "/workspace"
            command: {
                name: "./scripts/ci/tasks/release.sh"
            }
        }

        // Proto task using proto.sh
        proto: docker.#Run & {
            input: #BaseContainer
            workdir: "/workspace"
            command: {
                name: "./scripts/ci/tasks/proto.sh"
            }
        }

        // Dependencies task using dependencies.sh
        dependencies: docker.#Run & {
            input: #BaseContainer
            workdir: "/workspace"
            command: {
                name: "./scripts/ci/tasks/dependencies.sh"
            }
        }

        // Package task using package.sh
        package: docker.#Run & {
            input: #BaseContainer
            workdir: "/workspace"
            command: {
                name: "./scripts/ci/tasks/package.sh"
            }
        }

        // Utils
        utils: {
            // Health check using health-check.sh
            healthCheck: docker.#Run & {
                input: #BaseContainer
                workdir: "/workspace"
                command: {
                    name: "./scripts/ci/utils/health-check.sh"
                }
            }

            // Cleanup using cleanup.sh
            cleanup: docker.#Run & {
                input: #BaseContainer
                workdir: "/workspace"
                command: {
                    name: "./scripts/ci/utils/cleanup.sh"
                }
            }

            // Setup dev using setup-dev.sh
            setupDev: docker.#Run & {
                input: #BaseContainer
                workdir: "/workspace"
                command: {
                    name: "./scripts/ci/utils/setup-dev.sh"
                }
            }
        }

        // CI Pipeline
        ci: {
            pipeline: dagger.#Pipeline & {
                steps: [
                    utils.cleanup,
                    dependencies,
                    lint,
                    test,
                    build,
                    docker.build,
                    if client.env.CI_COMMIT_TAG != "" {
                        package
                    },
                ]
            }
        }
    }
}`

// Add new test templates after the existing templates
const testHelperTemplate = `#!/usr/bin/env bash
# Test helper functions and setup

# Load bats helpers
load '../../../node_modules/bats-support/load'
load '../../../node_modules/bats-assert/load'

# Set up test environment
setup() {
    # Create temp directory for tests
    TEST_TEMP_DIR="$(mktemp -d)"
    export TEST_TEMP_DIR
    
    # Set up mock project root
    export PROJECT_ROOT="${TEST_TEMP_DIR}/project"
    mkdir -p "${PROJECT_ROOT}"
    
    # Source common library
    source "${BATS_TEST_DIRNAME}/../../../lib/common.sh"
}

# Clean up after tests
teardown() {
    rm -rf "${TEST_TEMP_DIR}"
}

# Mock functions
mock_command() {
    local cmd=$1
    local exit_code=${2:-0}
    local output=${3:-""}
    
    eval "function ${cmd}() { echo \"${output}\"; return ${exit_code}; }"
    export -f "${cmd}"
}

# Create test files
create_test_file() {
    local path=$1
    local content=$2
    mkdir -p "$(dirname "${path}")"
    echo "${content}" > "${path}"
}
`

const commonTestTemplate = `#!/usr/bin/env bats
# Tests for common.sh library

load 'test_helper'

@test "is_ci returns true when CI environment variable is set" {
    export CI=true
    run is_ci
    assert_success
}

@test "is_ci returns false when CI environment variable is not set" {
    unset CI
    run is_ci
    assert_failure
}

@test "is_debug returns true when DEBUG environment variable is set" {
    export DEBUG=true
    run is_debug
    assert_success
}

@test "retry succeeds within allowed attempts" {
    mock_command "test_command" 1 "failed"
    mock_command "test_command" 0 "success"
    
    run retry 3 1 test_command
    
    assert_success
    assert_output --partial "success"
}

@test "retry fails after max attempts" {
    mock_command "test_command" 1 "failed"
    
    run retry 2 1 test_command
    
    assert_failure
    assert_output --partial "failed"
}

@test "ensure_command succeeds when command exists" {
    run ensure_command "bash"
    assert_success
}

@test "ensure_command fails when command doesn't exist" {
    run ensure_command "nonexistent_command"
    assert_failure
    assert_output --partial "Required command not found"
}

@test "load_env loads local environment file" {
    create_test_file "${PROJECT_ROOT}/.env" "TEST_VAR=local_value"
    
    run load_env
    
    assert_success
    assert [ "${TEST_VAR}" = "local_value" ]
}

@test "load_env loads environment-specific file when local doesn't exist" {
    create_test_file "${PROJECT_ROOT}/scripts/env/development.env" "TEST_VAR=dev_value"
    
    run load_env "development"
    
    assert_success
    assert [ "${TEST_VAR}" = "dev_value" ]
}
`

const loggerTestTemplate = `#!/usr/bin/env bats
# Tests for logger.sh library

load 'test_helper'
source "${BATS_TEST_DIRNAME}/../../../lib/logger.sh"

@test "log_debug is hidden when LOG_LEVEL is INFO" {
    export LOG_LEVEL=INFO
    run log_debug "test message"
    assert_output ""
}

@test "log_debug shows when LOG_LEVEL is DEBUG" {
    export LOG_LEVEL=DEBUG
    run log_debug "test message"
    assert_output --partial "DEBUG"
    assert_output --partial "test message"
}

@test "log_info shows when LOG_LEVEL is INFO" {
    export LOG_LEVEL=INFO
    run log_info "test message"
    assert_output --partial "INFO"
    assert_output --partial "test message"
}

@test "log_warn shows when LOG_LEVEL is WARN" {
    export LOG_LEVEL=WARN
    run log_warn "test message"
    assert_output --partial "WARN"
    assert_output --partial "test message"
}

@test "log_error shows when LOG_LEVEL is ERROR" {
    export LOG_LEVEL=ERROR
    run log_error "test message"
    assert_output --partial "ERROR"
    assert_output --partial "test message"
}

@test "log_fatal exits with status 1" {
    run log_fatal "test message"
    assert_failure
    assert_output --partial "FATAL"
    assert_output --partial "test message"
}

@test "progress_bar shows correct percentage" {
    run progress_bar 5 10 20
    assert_output --partial "50%"
}
`

const dockerTestTemplate = `#!/usr/bin/env bats
# Tests for docker.sh library

load 'test_helper'
source "${BATS_TEST_DIRNAME}/../../../lib/docker.sh"

@test "docker_build constructs correct build command" {
    mock_command "docker" 0 ""
    export CI_COMMIT_SHA="test_sha"
    
    run docker_build "test-image" "Dockerfile" "."
    
    assert_success
    assert_output --partial "Building Docker image: test-image"
}

@test "docker_push handles release tags correctly" {
    mock_command "docker" 0 ""
    export CI_COMMIT_TAG="v1.0.0"
    
    run docker_push "test-image" "registry.example.com"
    
    assert_success
    assert_output --partial "Pushing Docker image: registry.example.com/test-image"
}

@test "docker_cleanup removes unused resources" {
    mock_command "docker" 0 ""
    
    run docker_cleanup
    
    assert_success
    assert_output --partial "Cleaning up Docker resources"
}
`
