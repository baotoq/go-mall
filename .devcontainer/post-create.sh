#!/usr/bin/env bash
set -euo pipefail

# Docker creates named-volume mount points as root when the in-image path does
# not exist. Fix ownership of the mount point only — new files written by
# vscode inherit correct ownership, so recursive chown is unnecessary and slow
# on warm Go module caches.
for dir in "$HOME/go" "web/node_modules"; do
  mkdir -p "$dir"
  if [ "$(stat -c '%u' "$dir")" != "$(id -u)" ]; then
    sudo chown "$(id -u):$(id -g)" "$dir"
  fi
done

# Install Go protoc plugins + Wire + ent CLI (defined in Makefile `init` target)
make init
