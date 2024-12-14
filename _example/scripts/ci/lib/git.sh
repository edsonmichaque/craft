#!/usr/bin/env bash
# Git utility functions for version management and repository operations
# Provides functions for semantic versioning, tagging, and repository status

source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

# Version pattern validation
readonly VERSION_PATTERN='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-((0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(\.(0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*))?(\+([0-9a-zA-Z-]+(\.[0-9a-zA-Z-]+)*))?$'

# Ensure we're in a git repository
ensure_git_repo() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "Not a git repository"
        exit 1
    fi
}

# Get the current version from git tags with fallback strategies
get_version() {
    local version

    ensure_git_repo
    
    if [[ -n "${CI_COMMIT_TAG:-}" ]]; then
        version="${CI_COMMIT_TAG}"
    elif [[ -n "${VERSION:-}" ]]; then
        version="${VERSION}"
    else
        version=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
        
        # Add branch name for non-release versions
        if [[ "$version" == "dev" ]]; then
            local branch
            branch=$(get_branch_name)
            version="dev-${branch}-$(get_short_commit_hash)"
        fi
    fi
    
    echo "$version"
}

# Get the latest tag with validation
get_latest_tag() {
    ensure_git_repo
    
    local latest_tag
    latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    
    # Validate tag format
    if ! echo "$latest_tag" | grep -qE "$VERSION_PATTERN"; then
        log_warn "Latest tag '$latest_tag' does not follow semantic versioning"
        latest_tag="v0.0.0"
    fi
    
    echo "$latest_tag"
}

# Calculate the next version based on semver rules
get_next_version() {
    local current_version
    local next_version
    local bump_type=${1:-patch}
    local pre_release=${2:-}
    
    current_version=$(get_latest_tag)
    current_version=${current_version#v}
    
    # Split version into components
    if ! IFS='.' read -r major minor patch prerelease <<< "${current_version/-/ }"; then
        log_error "Failed to parse version: $current_version"
        exit 1
    fi
    
    # Remove any build metadata
    patch=${patch%%+*}
    
    case "$bump_type" in
        major)
            next_version="$((major + 1)).0.0"
            ;;
        minor)
            next_version="${major}.$((minor + 1)).0"
            ;;
        patch)
            next_version="${major}.${minor}.$((patch + 1))"
            ;;
        *)
            log_error "Invalid bump type: $bump_type (expected: major|minor|patch)"
            exit 1
            ;;
    esac
    
    # Add pre-release suffix if specified
    if [[ -n "$pre_release" ]]; then
        next_version="${next_version}-${pre_release}"
    fi
    
    echo "v${next_version}"
}

# Create and push a new tag with validation
create_tag() {
    local version=$1
    local message=${2:-"Release ${version}"}
    local force=${3:-false}
    
    ensure_git_repo
    
    # Validate version format
    if ! echo "$version" | grep -qE "$VERSION_PATTERN"; then
        log_error "Invalid version format: $version"
        log_error "Version must match pattern: $VERSION_PATTERN"
        exit 1
    }
    
    # Check if tag already exists
    if git rev-parse "$version" >/dev/null 2>&1; then
        if [[ "$force" != "true" ]]; then
            log_error "Tag $version already exists. Use force=true to override"
            exit 1
        fi
        log_warn "Force-updating existing tag: $version"
        git tag -d "$version"
        git push origin ":refs/tags/$version"
    fi
    
    log_info "Creating tag: $version"
    if ! git tag -a "$version" -m "$message"; then
        log_error "Failed to create tag: $version"
        exit 1
    fi
    
    if ! git push origin "$version"; then
        log_error "Failed to push tag: $version"
        git tag -d "$version"
        exit 1
    fi
}

# Generate a changelog between tags
get_changelog() {
    local from_tag=${1:-$(get_latest_tag)}
    local to_ref=${2:-HEAD}
    local format=${3:-"* %s (%h)"}
    
    ensure_git_repo
    
    if ! git rev-parse "$from_tag" >/dev/null 2>&1; then
        log_error "Invalid from_tag: $from_tag"
        exit 1
    fi
    
    if ! git rev-parse "$to_ref" >/dev/null 2>&1; then
        log_error "Invalid to_ref: $to_ref"
        exit 1
    fi
    
    git log "${from_tag}..${to_ref}" --no-merges --pretty=format:"$format"
}

# Repository status functions
is_working_directory_clean() {
    ensure_git_repo
    [[ -z "$(git status --porcelain)" ]]
}

is_current_branch_main() {
    local branch
    branch=$(get_branch_name)
    [[ "$branch" == "main" || "$branch" == "master" ]]
}

has_uncommitted_changes() {
    ! is_working_directory_clean
}

# Repository information functions
get_branch_name() {
    ensure_git_repo
    git rev-parse --abbrev-ref HEAD
}

get_commit_hash() {
    ensure_git_repo
    git rev-parse HEAD
}

get_short_commit_hash() {
    ensure_git_repo
    git rev-parse --short HEAD
}

get_repo_root() {
    ensure_git_repo
    git rev-parse --show-toplevel
}

get_repo_name() {
    ensure_git_repo
    basename "$(get_repo_root)"
}

# Utility functions
sync_tags() {
    ensure_git_repo
    log_info "Syncing tags with remote"
    git fetch --tags --force
    git push --tags
}

cleanup_old_tags() {
    local keep_count=${1:-10}
    ensure_git_repo
    
    log_info "Cleaning up old tags (keeping last $keep_count)"
    local tags
    tags=$(git tag --sort=-creatordate | tail -n +$((keep_count + 1)))
    
    if [[ -n "$tags" ]]; then
        echo "$tags" | xargs -r git tag -d
        echo "$tags" | xargs -r git push origin --delete
    fi
}
