.PHONY: generate build tidy

generate-ent:
	cd src/product/ent && go generate ./...

build:
	go build ./...

tidy:
	go mod tidy
