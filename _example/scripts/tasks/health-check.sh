#!/usr/bin/env bash
# Health check script
# Verifies the application is running correctly

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

check_process() {
    local binary=$1
    local pid_file="/var/run/${binary}.pid"

    if [[ -f "$pid_file" ]]; then
        local pid
        pid=$(cat "$pid_file")
        if kill -0 "$pid" 2>/dev/null; then
            log_info "Process is running (PID: $pid)"
            return 0
        fi
    fi

    log_error "Process is not running"
        return 1
}

check_http_endpoint() {
    local url=$1
    local expected_status=${2:-200}
    
    local response
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url")
    
    if [[ "$response" == "$expected_status" ]]; then
        log_info "HTTP endpoint is healthy"
        return 0
    fi
    
    log_error "HTTP endpoint returned status $response (expected $expected_status)"
        return 1
}

main() {
    local binary="craft"
    local health_url="http://localhost:8080/health"
    
    # Check process
    check_process "$binary"
    
    # Check HTTP endpoint
    check_http_endpoint "$health_url"
}

main "$@"