#!/bin/sh
set -e

if [ "$DEBUG" = "1" ]; then
    exec /app/dlv exec /app/app \
        --listen=:40000 \
        --headless \
        --api-version=2 \
        --accept-multiclient \
        --continue \
        -- -f etc/${SERVICE}-api.yaml
else
    exec /app/app -f etc/${SERVICE}-api.yaml
fi
