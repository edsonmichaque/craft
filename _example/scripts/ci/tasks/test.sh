#!/usr/bin/env bash
# Test task script
# Runs all project tests

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

run_tests() {
    local coverage_dir="${PROJECT_ROOT}/coverage"
    mkdir -p "$coverage_dir"
    
    log_info "Running tests with coverage"
    go test -race -coverprofile="${coverage_dir}/coverage.out" -covermode=atomic ./...
    
    if command -v go-junit-report >/dev/null 2>&1; then
        go test -v ./... 2>&1 | go-junit-report > "${coverage_dir}/report.xml"
    fi
    
    if command -v gocov >/dev/null 2>&1; then
        gocov convert "${coverage_dir}/coverage.out" | gocov-html > "${coverage_dir}/coverage.html"
    fi
}

main() {
    log_info "Starting test suite"
    
    # Verify required tools
    ensure_command "go"
    
    run_tests
    
    log_info "Tests completed successfully!"
}

main "$@"
