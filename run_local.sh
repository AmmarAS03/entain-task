#!/bin/sh
set -e

clear
cd core/cmd/core
go build

export APP_LOG_LEVEL=debug

docker-compose up -d

./core
