#!/bin/sh
set -e

min_protoc_version=32

protoc_version=$(protoc --version | awk '{print $2}')
protoc_major_version=${protoc_version%%.*}

if ((protoc_major_version < min_protoc_version)); then
  echo "protoc ${protoc_version} does not meet minimum required version of ${min_protoc_version}. Update protoc and try again."
  exit 1
fi

protoc -I . -I .. --go_opt=default_api_level=API_HYBRID --go_out=paths=source_relative:. --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:. core.proto
rm core_protoopaque.pb.go