.PHONY: generate-ent catalog-ent cart-ent payment-ent build tidy catalog-api cart-api payment-api

generate-ent: catalog-ent cart-ent payment-ent

catalog-ent:
	cd src/catalog/ent && go generate ./...

cart-ent:
	cd src/cart/ent && go generate ./...

payment-ent:
	cd src/payment/ent && go generate ./...

build:
	cd src/catalog && go build ./...
	cd src/cart && go build ./...
	cd src/payment && go build ./...

tidy:
	cd src/catalog && go mod tidy
	cd src/cart && go mod tidy
	cd src/payment && go mod tidy

catalog-api:
	cd src/catalog && go run product.go -f etc/catalog-api.yaml

cart-api:
	cd src/cart && go run cart.go -f etc/cart-api.yaml

payment-api:
	cd src/payment && go run payment.go -f etc/payment-api.yaml
