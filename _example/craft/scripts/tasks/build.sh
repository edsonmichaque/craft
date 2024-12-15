#!/usr/bin/env bash
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
    build_binary "craft" "${build_flags[@]}"

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