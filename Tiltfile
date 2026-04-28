# go-mall - Tiltfile
# Each service runs with a Dapr sidecar using project-local components.
# Components are in .dapr/components/ (pubsub, statestore, etc.)
# Keycloak auth server runs via Docker Compose.

docker_compose('deploy/docker-compose.keycloak.yml')

local_resource(
    'catalog-api',
    serve_cmd='dapr run --app-id catalog-api --app-port 9001 --dapr-http-port 3500 --resources-path ../../.dapr/components -- go run . -f etc/catalog-api.yaml',
    serve_dir='src/catalog',
    deps=[
        'src/catalog/main.go',
        'src/catalog/seed.go',
        'src/catalog/internal/',
        'src/catalog/etc/catalog-api.yaml',
        '.dapr/components/',
    ],
    labels=['service'],
)

local_resource(
    'cart-api',
    serve_cmd='dapr run --app-id cart-api --app-port 9002 --dapr-http-port 3501 --resources-path ../../.dapr/components -- go run cart.go -f etc/cart-api.yaml',
    serve_dir='src/cart',
    deps=[
        'src/cart/cart.go',
        'src/cart/internal/',
        'src/cart/etc/cart-api.yaml',
        '.dapr/components/',
    ],
    labels=['service'],
)

local_resource(
    'payment-api',
    serve_cmd='dapr run --app-id payment-api --app-port 9003 --dapr-http-port 3502 --resources-path ../../.dapr/components -- go run payment.go -f etc/payment-api.yaml',
    serve_dir='src/payment',
    deps=[
        'src/payment/payment.go',
        'src/payment/internal/',
        'src/payment/etc/payment-api.yaml',
        '.dapr/components/',
    ],
    labels=['service'],
)
