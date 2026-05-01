#!/usr/bin/env bash
set -euo pipefail

# Install Tilt via Homebrew
brew install tilt-dev/tap/tilt

# Install Go protoc plugins + Wire + ent CLI (defined in Makefile `init` target)
make init
