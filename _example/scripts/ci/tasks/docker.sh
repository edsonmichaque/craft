#!/usr/bin/env bash
# Docker task script
# Handles Docker image building and publishing

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"
source "$(dirname "${BASH_SOURCE[0]}")/../lib/docker.sh"

build_images() {
    for binary in "craftd" "craftctl" ; do
        docker_build "${CI_REGISTRY_IMAGE}/${binary}" "docker/${binary}.Dockerfile"
    done
}

push_images() {
    for binary in "craftd" "craftctl" ; do
        docker_push "${CI_REGISTRY_IMAGE}/${binary}"
    done
}

main() {
    log_info "Starting Docker build process"
    
    # Verify required tools
    ensure_command "docker"
    
    build_images
    
    if [[ -n "${CI:-}" ]]; then
        push_images
    fi
    
    log_info "Docker build complete!"
}

main "$@"
