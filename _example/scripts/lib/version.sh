#!/usr/bin/env bash
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
    local version_regex="^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$"
    
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