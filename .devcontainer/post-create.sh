#!/usr/bin/env bash
set -euo pipefail

# Install Go protoc plugins + Wire + ent CLI (defined in Makefile `init` target)
make init

# Install protoc, Dapr CLI, and Tilt via Homebrew
brew install protobuf dapr/tap/dapr tilt-dev/tap/tilt
