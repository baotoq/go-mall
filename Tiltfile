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
