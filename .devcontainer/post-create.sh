#!/usr/bin/env bash
set -euo pipefail

# Install protoc, Dapr CLI, and Tilt via Homebrew (protoc must precede make init)
brew install protobuf tilt-dev/tap/tilt

# Install Go protoc plugins + Wire + ent CLI (defined in Makefile `init` target)
make init
