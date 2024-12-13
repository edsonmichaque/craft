#!/usr/bin/env bash
# Docker utility functions

source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

# Docker build with caching and multi-stage optimization
docker_build() {
    local image=$1
    local dockerfile=$2
    local context=${3:-.}
    local cache_from=""
    local build_args=()

    # Add build arguments
    build_args+=(--build-arg "VERSION=${CI_COMMIT_TAG:-dev}")
    build_args+=(--build-arg "COMMIT=${CI_COMMIT_SHA}")
    build_args+=(--build-arg "BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")")

    # Use cache from previous builds if available
    if [[ -n "${CI_REGISTRY_IMAGE:-}" ]]; then
        cache_from="--cache-from ${CI_REGISTRY_IMAGE}:${CI_COMMIT_BRANCH:-main}"
    fi

    log_info "Building Docker image: $image"
    docker build \
        "${build_args[@]}" \
        $cache_from \
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

# Clean up old images and containers
docker_cleanup() {
    log_info "Cleaning up Docker resources"
    
    # Remove stopped containers
    docker container prune -f
    
    # Remove unused images
    docker image prune -f
    
    # Remove unused volumes
    docker volume prune -f
}
