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