#!/usr/bin/env bash
# Lint task script
# Runs all linters and code quality checks

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

run_linters() {
    log_info "Running golangci-lint"
    golangci-lint run ./...
    
    log_info "Running go vet"
    go vet ./...
    
    log_info "Checking go fmt"
    if [ -n "$(gofmt -l .)" ]; then
        log_error "Code is not formatted. Please run 'go fmt ./...'"
        return 1
    fi
}

main() {
    log_info "Starting code quality checks"
    
    # Verify required tools
    ensure_command "golangci-lint"
    ensure_command "go"
    
    run_linters
    
    log_info "Code quality checks passed!"
}

main "$@"
