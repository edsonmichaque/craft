#!/usr/bin/env bash
# Version management functions

source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

# Get the current version
get_version() {
    if [[ -n "${CI_COMMIT_TAG:-}" ]]; then
        echo "$CI_COMMIT_TAG"
    else
        echo "dev"
    fi
}

# Get the current commit hash
get_commit() {
    git rev-parse HEAD
}

# Get the current commit short hash
get_commit_short() {
    git rev-parse --short HEAD
}

# Bump version according to semver
bump_version() {
    local current_version=$1
    local bump_type=${2:-patch}
    
    local major minor patch
    IFS='.' read -r major minor patch <<< "${current_version#v}"
    
    case "$bump_type" in
        major) echo "v$((major + 1)).0.0" ;;
        minor) echo "v${major}.$((minor + 1)).0" ;;
        patch) echo "v${major}.${minor}.$((patch + 1))" ;;
        *) echo "$current_version" ;;
    esac
}
