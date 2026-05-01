#!/usr/bin/env bash
set -euo pipefail

# Install protoc (required for `make api` and `make config`)
sudo apt-get update -qq
sudo apt-get install -y --no-install-recommends protobuf-compiler

# Install Go protoc plugins + Wire + ent CLI (defined in Makefile `init` target)
make init

# Install Dapr CLI (sidecar managed separately via Helm/Tilt, but CLI is useful locally)
wget -q https://raw.githubusercontent.com/dapr/cli/master/install/install.sh -O - | /bin/bash

# Install Tilt (used by `make dev` / `make debug`)
curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash
