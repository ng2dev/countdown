.PHONY: all clean install test tf cover protofmt protoc protolint protodocs import-spec

# make sure we turn on go modules
export GO111MODULE := on

TOOLS := cmd/countdown cmd/countdowncli

# MODE=count records heat map in test coverage
# MODE=set just records which lines were hit by one test
MODE ?= set

# Check if linter exists
LINT := $(shell command -v golangci-lint 2> /dev/null)

# for dockerized prototool
USER := $(shell id -u):$(shell id -g)
DOCKER_BASE := docker run --rm --user=${USER} -v $(shell pwd):/work iov1/prototool:v0.2.2
PROTOTOOL := $(DOCKER_BASE) prototool
PROTOC := $(DOCKER_BASE) protoc
WEAVEDIR=$(shell go list -m -f '{{.Dir}}' github.com/iov-one/weave)

all: import-spec test lint install

dist:
	cd cmd/countdown && $(MAKE) dist

install:
	for ex in $(TOOLS); do cd $$ex && make install && cd -; done

test:
	@# countdown binary is required by some tests. In order to not skip them, ensure countdown binary is provided and in the latest version.
	go vet -mod=readonly ./...
	go test -mod=readonly -race ./...

lint:
	@go mod vendor
	docker run --rm -it -v $(shell pwd):/go/src/github.com/ng2dev/countdown -w="/go/src/github.com/ng2dev/countdown" golangci/golangci-lint:v1.17.1 golangci-lint run ./...
	@rm -rf vendor

# Test fast
tf:
	go test -short ./...

test-verbose:
	go vet ./...
	go test -v -race ./...

mod:
	go mod tidy

cover:
	@go test -mod=readonly -covermode=$(MODE) \
		-coverpkg=github.com/ng2dev/countdown/cmd/countdown/app, \
		-coverprofile=coverage/countdownd_app.out \
		github.com/ng2dev/countdown/cmd/countdown/app
		cat coverage/*.out > coverage/coverage.txt
	@go test -mod=readonly -covermode=$(MODE) \
		-coverpkg=github.com/ng2dev/countdown/cmd/countdown/app,github.com/ng2dev/countdown/cmd/countdown/client \
		-coverprofile=coverage/countdownd_client.out \
		github.com/ng2dev/countdown/cmd/countdown/client

novendor:
	@rm -rf ./vendor

protolint: novendor
	$(PROTOTOOL) lint

protofmt: novendor
	$(PROTOTOOL) format -w

protodocs: novendor
	./scripts/build_protodocs.sh docs/proto

protoc: protolint protofmt protodocs
	$(PROTOTOOL) generate

import-spec:
	@rm -rf ./spec
	@mkdir -p spec/github.com/iov-one/weave
	@cp -r ${WEAVEDIR}/spec/gogo/* spec/github.com/iov-one/weave
	@chmod -R +w spec

inittm:
	tendermint init --home ~/.countdown

runtm:
	tendermint node --home ~/.countdown > ~/.countdown/tendermint.log &
