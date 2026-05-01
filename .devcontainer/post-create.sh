#!/usr/bin/env bash
set -euo pipefail

# Install protoc (required for `make api` and `make config`)
sudo apt-get update -qq
sudo apt-get install -y --no-install-recommends protobuf-compiler

# Install Go protoc plugins + Wire + ent CLI (defined in Makefile `init` target)
make init

# Install Dapr CLI and Tilt via Homebrew
brew install dapr/tap/dapr tilt-dev/tap/tilt
