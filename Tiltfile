# go-mall - Tiltfile

local_resource(
    'catalog-api',
    serve_cmd='go run product.go -f etc/catalog-api.yaml',
    serve_dir='src/catalog',
    deps=[
        'src/catalog/product.go',
        'src/catalog/internal/',
        'src/catalog/etc/catalog-api.yaml',
    ],
    labels=['service'],
)

local_resource(
    'cart-api',
    serve_cmd='go run cart.go -f etc/cart-api.yaml',
    serve_dir='src/cart',
    deps=[
        'src/cart/cart.go',
        'src/cart/internal/',
        'src/cart/etc/cart-api.yaml',
    ],
    labels=['service'],
)

local_resource(
    'payment-api',
    serve_cmd='go run payment.go -f etc/payment-api.yaml',
    serve_dir='src/payment',
    deps=[
        'src/payment/payment.go',
        'src/payment/internal/',
        'src/payment/etc/payment-api.yaml',
    ],
    labels=['service'],
)
