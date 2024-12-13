#!/usr/bin/env bash
# Proto task script
# Handles protobuf compilation

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

generate_protos() {
    local proto_dir="${PROJECT_ROOT}/proto"
    
    if [[ ! -d "$proto_dir" ]]; then
        log_info "No proto directory found, skipping"
        return 0
    }
    
    log_info "Generating protobuf code"
    protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        "${proto_dir}"/*.proto
}

main() {
    log_info "Starting protobuf generation"
    
    # Verify required tools
    ensure_command "protoc"
    ensure_command "protoc-gen-go"
    ensure_command "protoc-gen-go-grpc"
    
    generate_protos
    
    log_info "Protobuf generation complete!"
}

main "$@"
