#!/usr/bin/env bash
# Build task script
# Handles building all project binaries

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

# Build configuration
BUILD_DIR="${PROJECT_ROOT}/bin"
DIST_DIR="${PROJECT_ROOT}/dist"
PLATFORMS=("linux/amd64" "darwin/amd64" "windows/amd64")
CGO_ENABLED=0

build_binary() {
    local binary=$1
    local os=$2
    local arch=$3
    local output
    
    if [[ $os == "windows" ]]; then
        output="${BUILD_DIR}/${binary}-${os}-${arch}.exe"
    else
        output="${BUILD_DIR}/${binary}-${os}-${arch}"
    fi

    log_info "Building ${binary} for ${os}/${arch}"
    
    GOOS=$os GOARCH=$arch CGO_ENABLED=$CGO_ENABLED \
    go build -ldflags "${LDFLAGS}" \
        -o "$output" \
        "${PROJECT_ROOT}/cmd/${binary}"
}

build_all() {
    mkdir -p "$BUILD_DIR"
    
    for binary in "craftd" "craftctl" ; do
        for platform in "${PLATFORMS[@]}"; do
            IFS='/' read -r os arch <<< "$platform"
            build_binary "$binary" "$os" "$arch"
        done
    done
}

package_artifacts() {
    mkdir -p "$DIST_DIR"
    
    for binary in "craftd" "craftctl" ; do
        for platform in "${PLATFORMS[@]}"; do
            IFS='/' read -r os arch <<< "$platform"
            local src="${BUILD_DIR}/${binary}-${os}-${arch}"
            local dst="${DIST_DIR}/${binary}-${os}-${arch}"
            
            if [[ $os == "windows" ]]; then
                src="${src}.exe"
                dst="${dst}.exe"
            fi
            
            cp "$src" "$dst"
            
            # Create checksums
            (cd "$DIST_DIR" && sha256sum "$(basename "$dst")" > "$(basename "$dst").sha256")
        done
    done
}

main() {
    log_info "Starting build process"
    
    # Verify required tools
    ensure_command "go"
    
    # Set version information
    VERSION=$(get_version)
    COMMIT=$(get_commit)
    DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    # Set build flags
    LDFLAGS="-s -w \
        -X 'github.com/edsonmichaque/craft/pkg/version.Version=${VERSION}' \
        -X 'github.com/edsonmichaque/craft/pkg/version.GitCommit=${COMMIT}' \
        -X 'github.com/edsonmichaque/craft/pkg/version.BuildTime=${DATE}'"
    
    build_all
    package_artifacts
    
    log_info "Build complete!"
}

main "$@"
