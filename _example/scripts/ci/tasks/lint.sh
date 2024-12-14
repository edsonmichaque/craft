#!/usr/bin/env bash
# Lint task script
# Runs linters and code quality checks

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

main() {
    # Install golangci-lint if not present
    ensure_golangci_lint
    
    log_info "Running linters..."
    
    # Run golangci-lint
    golangci-lint run \
        --timeout=5m \
        --config="${PROJECT_ROOT}/.golangci.yml" \
        ./...
    
    # Run go fmt
    check_formatting
    
    # Run go vet
    go vet ./...
    
    log_info "Linting complete!"
}

ensure_golangci_lint() {
    if ! command -v golangci-lint >/dev/null 2>&1; then
        log_info "Installing golangci-lint..."
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    fi
}

check_formatting() {
    log_info "Checking code formatting..."
    
    local files
    files=$(gofmt -l .)
    
    if [[ -n "$files" ]]; then
        log_error "The following files are not properly formatted:"
        echo "$files"
        exit 1
    fi
}

main "$@"
