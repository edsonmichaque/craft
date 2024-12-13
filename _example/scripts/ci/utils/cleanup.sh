#!/usr/bin/env bash
# Cleanup utility
# Removes temporary files and artifacts

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

cleanup_build() {
    log_info "Cleaning build artifacts"
    rm -rf "${PROJECT_ROOT}/bin"
    rm -rf "${PROJECT_ROOT}/dist"
}

cleanup_deps() {
    log_info "Cleaning dependency cache"
    go clean -cache -modcache -i -r
}

cleanup_docker() {
    log_info "Cleaning Docker resources"
    source "$(dirname "${BASH_SOURCE[0]}")/../lib/docker.sh"
    docker_cleanup
}

main() {
    log_info "Starting cleanup process"
    
    cleanup_build
    
    if [[ "${1:-}" == "--deep" ]]; then
        cleanup_deps
        cleanup_docker
    fi
    
    log_info "Cleanup complete!"
}

main "$@"
