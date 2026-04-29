# go-mall - Tiltfile
# Each service runs with a Dapr sidecar using project-local components.
# Components are in .dapr/components/ (pubsub, statestore, etc.)
# Keycloak auth server runs via Docker Compose.

docker_compose('deploy/docker-compose.keycloak.yml')
docker_compose('deploy/docker-compose.observability.yml')

debug = os.getenv('DEBUG', 'false') == 'true'

def go_cmd(main, port, args):
    if debug:
        return 'dlv debug --headless --listen=:{port} --api-version=2 --accept-multiclient {main} -- {args}'.format(
            port=port, main=main, args=args)
    return 'go run {main} {args}'.format(main=main, args=args)

local_resource(
    'catalog-api',
    serve_cmd='dapr run --app-id catalog-api --app-port 9001 --app-protocol http --dapr-http-port 9901 --resources-path ../../.dapr/components --config ../../.dapr/config.yaml -- ' + go_cmd('.', 40001, '-f etc/catalog-api.yaml'),
    serve_dir='src/catalog',
    deps=[
        'src/catalog/main.go',
        'src/catalog/seed.go',
        'src/catalog/internal/',
        'src/catalog/etc/catalog-api.yaml',
        '.dapr/components/',
        '.dapr/config.yaml',
    ],
    labels=['service'],
)

local_resource(
    'cart-api',
    serve_cmd='dapr run --app-id cart-api --app-port 9002 --app-protocol http --dapr-http-port 9902 --resources-path ../../.dapr/components --config ../../.dapr/config.yaml -- ' + go_cmd('.', 40002, '-f etc/cart-api.yaml'),
    serve_dir='src/cart',
    deps=[
        'src/cart/cart.go',
        'src/cart/internal/',
        'src/cart/etc/cart-api.yaml',
        '.dapr/components/',
        '.dapr/config.yaml',
    ],
    labels=['service'],
)

local_resource(
    'payment-api',
    serve_cmd='dapr run --app-id payment-api --app-port 9003 --app-protocol http --dapr-http-port 9903 --resources-path ../../.dapr/components --config ../../.dapr/config.yaml -- ' + go_cmd('.', 40003, '-f etc/payment-api.yaml'),
    serve_dir='src/payment',
    deps=[
        'src/payment/payment.go',
        'src/payment/internal/',
        'src/payment/etc/payment-api.yaml',
        '.dapr/components/',
        '.dapr/config.yaml',
    ],
    labels=['service'],
)
