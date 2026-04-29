.PHONY: generate-ent catalog-ent cart-ent payment-ent build tidy catalog-api cart-api payment-api debug-catalog debug-cart debug-payment debug-all

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

up-debug:
	tilt up -- --debug=catalog --debug=cart --debug=payment
