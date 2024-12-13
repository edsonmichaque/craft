#!/usr/bin/env bash
# Git utility functions

source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

# Get the current git branch
get_branch() {
    git rev-parse --abbrev-ref HEAD
}

# Get the latest git tag
get_latest_tag() {
    git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"
}

# Check if working directory is clean
is_working_directory_clean() {
    [[ -z "$(git status --porcelain)" ]]
}

# Get all changes since last tag
get_changelog() {
    local latest_tag
    latest_tag=$(get_latest_tag)
    git log "${latest_tag}..HEAD" --pretty=format:"- %s" --no-merges
}

# Tag the current commit
tag_version() {
    local version=$1
    local message=${2:-"Release ${version}"}
    
    if ! is_working_directory_clean; then
        log_error "Working directory is not clean"
        return 1
    fi
    
    git tag -a "$version" -m "$message"
    git push origin "$version"
}
