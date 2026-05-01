#!/usr/bin/env bash
set -euo pipefail

# Install Go protoc plugins + Wire + ent CLI (defined in Makefile `init` target)
make init

# Install frontend dependencies into the named volume (web/node_modules)
cd web && npm ci
