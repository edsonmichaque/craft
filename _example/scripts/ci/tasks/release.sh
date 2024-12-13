#!/usr/bin/env bash
# Release task script
# Handles versioning and release process

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

create_release() {
    local version=$1
    local branch=${2:-main}
    
    log_info "Creating release ${version} from branch ${branch}"
    
    # Ensure we're on the correct branch
    git checkout "$branch"
    git pull origin "$branch"
    
    # Tag the release
    tag_version "$version"
    
    # Build and package
    "${PROJECT_ROOT}/scripts/ci/tasks/build.sh"
    
    # Create GitHub release if gh CLI is available
    if command -v gh >/dev/null 2>&1; then
        gh release create "$version" \
            --title "Release ${version}" \
            --notes "$(get_changelog)" \
            ./dist/*
    fi
}

main() {
    local version=${1:-}
    local branch=${2:-main}
    
    if [[ -z "$version" ]]; then
        log_error "Version parameter is required"
        exit 1
    fi
    
    create_release "$version" "$branch"
    
    log_info "Release ${version} created successfully!"
}

main "$@"
