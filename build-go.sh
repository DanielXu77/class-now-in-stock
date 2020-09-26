#!/bin/bash

# Clean up dependencies
# go mod tidy

# Create go_bin folder to place the build binary
mkdir -p go_bin

# Build go binary
# unused flags: -race
CGO_ENABLED=0 go build -ldflags "-extldflags '-static'" -o ./go_bin/cnis ./... && echo "Built successfully"