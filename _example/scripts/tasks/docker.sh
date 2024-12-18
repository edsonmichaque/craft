#!/usr/bin/env bash
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
    docker_build "craft:${version}" "$dockerfile" "$context"
    
    # Build any additional images (e.g., debug, minimal)
    if [[ -f "${PROJECT_ROOT}/build/docker/Dockerfile.debug" ]]; then
        docker_build "craft:${version}-debug" \
            "${PROJECT_ROOT}/build/docker/Dockerfile.debug" \
            "$context"
    fi
}

push_images() {
    local version=$1
    local registry=${DOCKER_REGISTRY:-""}
    
    # Push the main application image
    docker_push "craft:${version}" "$registry"
    
    # Push any additional images
    if [[ -f "${PROJECT_ROOT}/build/docker/Dockerfile.debug" ]]; then
        docker_push "craft:${version}-debug" "$registry"
    fi
}

main "$@"