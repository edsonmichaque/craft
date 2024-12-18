#!/usr/bin/env bash
# Release task script
# Handles version bumping and release creation

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"
source "$(dirname "${BASH_SOURCE[0]}")/../lib/git.sh"
source "$(dirname "${BASH_SOURCE[0]}")/../lib/version.sh"

main() {
    local bump_type=${1:-patch}
    local next_version
    
    # Ensure working directory is clean
    if ! is_working_directory_clean; then
        log_error "Working directory is not clean"
        exit 1
    }
    
    # Get next version
    next_version=$(get_next_version "$bump_type")
    
    # Validate version format
    if ! validate_version "$next_version"; then
        exit 1
    }
    
    # Generate changelog
    local changelog
    changelog=$(get_changelog)
    
    # Create and push tag
    create_tag "$next_version" "Release ${next_version}\n\n${changelog}"
    
    log_info "Release ${next_version} created successfully!"
    log_info "Changelog:\n${changelog}"
}

main "$@"
