GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)

ifeq ($(GOHOSTOS), windows)
    #the `find.exe` is different from `find` in bash/shell.
    #to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
    #changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
    #Git_Bash= $(subst cmd\,bin\bash.exe,$(dir $(shell where git)))
    Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
    INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find app -name *.proto -not -path '*/api/*.proto'")
    API_GREETER_PROTO_FILES=$(shell $(Git_Bash) -c "find api/greeter -name *.proto")
    API_CATALOG_PROTO_FILES=$(shell $(Git_Bash) -c "find api/catalog -name *.proto")
    API_CART_PROTO_FILES=$(shell $(Git_Bash) -c "find api/cart -name *.proto")
else
    INTERNAL_PROTO_FILES=$(shell find app -name *.proto -not -path "*/api/*.proto")
    API_GREETER_PROTO_FILES=$(shell find api/greeter -name *.proto)
    API_CATALOG_PROTO_FILES=$(shell find api/catalog -name *.proto)
    API_CART_PROTO_FILES=$(shell find api/cart -name *.proto)
endif

.PHONY: init
# init env
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/google/wire/cmd/wire@latest
	go install entgo.io/ent/cmd/ent@latest

.PHONY: config
# generate internal proto
config:
	protoc --proto_path=. \
	       --proto_path=./third_party \
	       --go_out=paths=source_relative:. \
	       $(INTERNAL_PROTO_FILES)

.PHONY: api-greeter
# generate greeter api proto
api-greeter:
	protoc --proto_path=./api \
	       --proto_path=./third_party \
	       --go_out=paths=source_relative:./api \
	       --go-http_out=paths=source_relative:./api \
	       --go-grpc_out=paths=source_relative:./api \
	       --openapi_out=fq_schema_naming=true,default_response=false:app/greeter \
	       $(API_GREETER_PROTO_FILES)

.PHONY: api-catalog
# generate catalog api proto
api-catalog:
	protoc --proto_path=./api \
	       --proto_path=./third_party \
	       --go_out=paths=source_relative:./api \
	       --go-http_out=paths=source_relative:./api \
	       --go-grpc_out=paths=source_relative:./api \
	       --openapi_out=fq_schema_naming=true,default_response=false:app/catalog \
	       $(API_CATALOG_PROTO_FILES)

.PHONY: api-cart
# generate cart api proto
api-cart:
	protoc --proto_path=./api \
	       --proto_path=./third_party \
	       --go_out=paths=source_relative:./api \
	       --go-http_out=paths=source_relative:./api \
	       --go-grpc_out=paths=source_relative:./api \
	       --openapi_out=fq_schema_naming=true,default_response=false:app/cart \
	       $(API_CART_PROTO_FILES)

.PHONY: api
# generate all api proto
api: api-greeter api-catalog api-cart

.PHONY: build-greeter
# build greeter
build-greeter:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/greeter ./app/greeter/cmd/server

.PHONY: build-catalog
# build catalog
build-catalog:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/catalog ./app/catalog/cmd/server

.PHONY: build
# build all services
build: build-greeter build-catalog

.PHONY: generate
# generate
generate:
	go generate ./...
	go mod tidy

.PHONY: dev
# start dev environment, Delve starts immediately
dev:
	tilt up -- --continue

.PHONY: debug
# start dev environment, wait for debugger to attach
debug:
	tilt up

.PHONY: all
# generate all
all:
	$(MAKE) api
	$(MAKE) config
	$(MAKE) generate

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
