.PHONY: all build test image tf protoc clean dist

BUILD_VERSION ?= $(shell git describe --tags)
BUILD_FLAGS := -mod=readonly -ldflags "-X github.com/iov-one/countdown/countdown.Version=${BUILD_VERSION}"
DOCKER_BUILD_FLAGS := -a -installsuffix cgo
BUILDOUT ?= countdown
IMAGE_NAME = "iov1/countdown:${BUILD_VERSION}"

all: dist

dist: clean test build image

clean:
	rm -f ${BUILDOUT}

build:
	GOARCH=amd64 CGO_ENABLED=0 GOOS=linux go build $(BUILD_FLAGS) $(DOCKER_BUILD_FLAGS) -o $(BUILDOUT) .

image:
	docker build --pull -t $(IMAGE_NAME) .

test:
	go test -mod=readonly -race ./...

install:
	go install $(BUILD_FLAGS) .

# Test fast
tf:
	go test -short ./...
