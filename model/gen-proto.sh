#!/usr/bin/env bash
set -e

min_protoc_version=32

protoc_version=$(protoc --version | awk '{print $2}')
protoc_major_version=${protoc_version%%.*}

if ((protoc_major_version < min_protoc_version)); then
  echo "protoc ${protoc_version} does not meet minimum required version of ${min_protoc_version}. Update protoc and try again."
  exit 1
fi

echo "Generating proto (Hybrid API)"
protoc -I . --go_opt=default_api_level=API_HYBRID --go_out=paths=source_relative:. --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:. event.proto
rm event_protoopaque.pb.go

go fmt ./...
go build
