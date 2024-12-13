#!/usr/bin/env bash
# Health check utility
# Verifies service health

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

check_service() {
    local host=${1:-localhost}
    local port=${2:-8080}
    local endpoint=${3:-/health}
    local timeout=${4:-5}
    
    curl --silent --fail \
        --max-time "$timeout" \
        "http://${host}:${port}${endpoint}"
}

main() {
    log_info "Running health checks"
    
    for binary in "craftd" "craftctl" ; do
        if ! check_service "localhost" "8080" "/health"; then
            log_error "Service ${binary} is not healthy"
            exit 1
        fi
    done
    
    log_info "All services are healthy!"
}

main "$@"
