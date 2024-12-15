#!/usr/bin/env bash
# Test task script
# Runs tests and generates coverage reports

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

main() {
    local coverage_dir="${PROJECT_ROOT}/coverage"
    mkdir -p "$coverage_dir"

    # Run tests with coverage
    run_tests "$coverage_dir"
    
    # Check coverage threshold
    check_coverage_threshold "$coverage_dir/coverage.out"
    
    # Generate coverage report
    generate_coverage_report "$coverage_dir"
}

run_tests() {
    local coverage_dir=$1
    
    log_info "Running tests..."
    
    go test \
        -race \
        -coverprofile="$coverage_dir/coverage.out" \
        -covermode=atomic \
        ./...
}

check_coverage_threshold() {
    local coverage_file=$1
    local threshold=${COVERAGE_THRESHOLD:-70}
    
    local coverage
    coverage=$(go tool cover -func="$coverage_file" | grep total: | awk '{print $3}' | sed 's/%//')
    
    log_info "Total coverage: ${coverage}%"
    
    if (( $(echo "$coverage < $threshold" | bc -l) )); then
        log_error "Coverage ${coverage}% is below threshold ${threshold}%"
        exit 1
    fi
}

generate_coverage_report() {
    local coverage_dir=$1
    
    log_info "Generating coverage report..."
    go tool cover -html="$coverage_dir/coverage.out" -o "$coverage_dir/coverage.html"
}

main "$@"