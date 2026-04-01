# go-mall - Tiltfile

local_resource(
    'product-api',
    serve_cmd='go run product.go -f etc/product-api.yaml',
    serve_dir='src/product',
    deps=[
        'src/product/product.go',
        'src/product/internal/',
        'src/product/etc/product-api.yaml',
    ],
    labels=['service'],
)
