.PHONY: generate build tidy

generate-ent:
	cd src/catalog/ent && go generate ./...

build:
	go build ./...

tidy:
	go mod tidy
