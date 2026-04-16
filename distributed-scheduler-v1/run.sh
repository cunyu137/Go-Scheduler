#!/usr/bin/env bash
set -e

docker compose up -d
sleep 8
go mod tidy
go run ./cmd/server
