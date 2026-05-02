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
    API_CATALOG_PROTO_FILES=$(shell $(Git_Bash) -c "find api/catalog -name *.proto")
    API_CART_PROTO_FILES=$(shell $(Git_Bash) -c "find api/cart -name *.proto")
    API_PAYMENT_PROTO_FILES=$(shell $(Git_Bash) -c "find api/payment -name *.proto")
    API_ORDER_PROTO_FILES=$(shell $(Git_Bash) -c "find api/order -name *.proto")
else
    INTERNAL_PROTO_FILES=$(shell find app -name *.proto -not -path "*/api/*.proto")
    API_CATALOG_PROTO_FILES=$(shell find api/catalog -name *.proto)
    API_CART_PROTO_FILES=$(shell find api/cart -name *.proto)
    API_PAYMENT_PROTO_FILES=$(shell find api/payment -name *.proto)
    API_ORDER_PROTO_FILES=$(shell find api/order -name *.proto)
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

.PHONY: api-payment
# generate payment api proto
api-payment:
	protoc --proto_path=./api \
	       --proto_path=./third_party \
	       --go_out=paths=source_relative:./api \
	       --go-http_out=paths=source_relative:./api \
	       --go-grpc_out=paths=source_relative:./api \
	       --openapi_out=fq_schema_naming=true,default_response=false:app/payment \
	       $(API_PAYMENT_PROTO_FILES)

.PHONY: api-order
# generate order api proto
api-order:
	protoc --proto_path=./api \
	       --proto_path=./third_party \
	       --go_out=paths=source_relative:./api \
	       --go-http_out=paths=source_relative:./api \
	       --go-grpc_out=paths=source_relative:./api \
	       --openapi_out=fq_schema_naming=true,default_response=false:app/order \
	       $(API_ORDER_PROTO_FILES)

.PHONY: api
# generate all api proto
api: api-catalog api-cart api-payment api-order

.PHONY: build-catalog
# build catalog
build-catalog:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/catalog ./app/catalog/cmd/server

.PHONY: build-payment
# build payment
build-payment:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/payment ./app/payment/cmd/server

.PHONY: build-order
# build order
build-order:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/order ./app/order/cmd/server

.PHONY: build-cart
# build cart
build-cart:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/cart ./app/cart/cmd/server

.PHONY: build
# build all services
build: build-catalog build-payment build-order build-cart

.PHONY: generate
# generate
generate:
	go generate ./...
	go mod tidy

.PHONY: verify-saga-config
# verify saga Dapr config AC9a + AC9c
verify-saga-config:
	@command -v yq >/dev/null 2>&1 || { echo "yq required (brew install yq or go install github.com/mikefarah/yq/v4@latest)"; exit 1; }
	@echo "AC9a: workflowstore actorStateStore=true"
	@yq -e '.spec.metadata[] | select(.name=="actorStateStore") | .value' deploy/k8s/base/infra/dapr/workflowstore.yaml | grep -q 'true' && echo PASS || { echo FAIL; exit 1; }
	@echo "AC9c: order-workflow-config has stateRetentionPolicy"
	@yq -e '.spec.workflow.stateRetentionPolicy' deploy/k8s/base/infra/dapr/workflow-config.yaml >/dev/null && echo PASS || { echo FAIL; exit 1; }
	@echo "AC9c: order Deployment annotated with dapr.io/config=order-workflow-config"
	@yq -e '.spec.template.metadata.annotations."dapr.io/config"' deploy/k8s/base/app/order.yaml | grep -q 'order-workflow-config' && echo PASS || { echo FAIL; exit 1; }

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
