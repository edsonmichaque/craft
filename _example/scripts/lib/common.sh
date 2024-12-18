#!/usr/bin/env bash
# Common library for scripts
# Provides core functionality used across all scripts

set -euo pipefail
IFS=$'\n\t'

# Script location
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Import other libraries
source "${SCRIPT_DIR}/logger.sh"
source "${SCRIPT_DIR}/version.sh"

# Global variables
export CI_COMMIT_SHA="${CI_COMMIT_SHA:-$(git rev-parse HEAD)}"
export CI_COMMIT_SHORT_SHA="${CI_COMMIT_SHORT_SHA:-$(git rev-parse --short HEAD)}"
export CI_COMMIT_BRANCH="${CI_COMMIT_BRANCH:-$(git rev-parse --abbrev-ref HEAD)}"
export CI_COMMIT_TAG="${CI_COMMIT_TAG:-$(git describe --tags --exact-match 2>/dev/null || echo '')}"
export CI_PROJECT_NAME="craft"

# Environment detection
is_ci() {
    [[ -n "${CI:-}" ]]
}

is_debug() {
    [[ -n "${DEBUG:-}" ]]
}

is_dry_run() {
    [[ -n "${DRY_RUN:-}" ]]
}

# Error handling
trap 'error_handler $?' ERR

error_handler() {
    local exit_code=$1
    log_error "Error occurred in $(caller) with exit code $exit_code"
    exit "$exit_code"
}

# Utility functions
retry() {
    local retries=${1:-3}
    local wait=${2:-5}
    local cmd=${@:3}
    local retry_count=0

    until [[ $retry_count -ge $retries ]]; do
        if eval "$cmd"; then
            return 0
        fi
        retry_count=$((retry_count + 1))
        log_warn "Command failed, attempt $retry_count/$retries"
        sleep "$wait"
    done
    return 1
}

ensure_command() {
    local cmd=$1
    if ! command -v "$cmd" >/dev/null 2>&1; then
        log_error "Required command not found: $cmd"
        exit 1
    fi
}

# Configuration management
load_env() {
    local env=${1:-development}
    local env_file="${PROJECT_ROOT}/scripts/env/${env}.env"
    local local_env_file="${PROJECT_ROOT}/.env"
    
    if [[ -f "$local_env_file" ]]; then
        log_info "Loading local environment from $local_env_file"
        set -o allexport
        source "$local_env_file"
        set +o allexport
    elif [[ -f "$env_file" ]]; then
        log_info "Loading environment from $env_file"
        set -o allexport
        source "$env_file"
        set +o allexport
    else
        log_warn "Environment file not found: $env_file"
    fi
}

# Cleanup handling
cleanup() {
    local exit_code=$?
    log_info "Cleaning up..."
    exit "$exit_code"
}
trap cleanup EXIT