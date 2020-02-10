.PHONY: build build-alpine clean test help default

BIN_NAME=bosun

VERSION := $(shell git describe --exact-match --tags 2> /dev/null || git describe --tags )
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_COMMIT_SHORT=$(shell git rev-parse --short HEAD)
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+dirty" || true)
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
IMAGE_NAME := sherifabdlnaby/bosun
FLAGS := -X github.com/sherifabdlnaby/bosun/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/sherifabdlnaby/bosun/version.Version=${VERSION} -X github.com/sherifabdlnaby/bosun/version.BuildDate=${BUILD_DATE}

default: run

help:
	@echo 'Management commands for bosun:'
	@echo
	@echo 'Usage:'
	@echo '    make build           Compile the project.'
	@echo '    make run		        Compile and run the project.'
	@echo '    make build-alpine    Compile optimized for alpine linux.'
	@echo '    make package         Build final docker image with just the go binary inside'
	@echo '    make tag             Tag image created by package with latest, git commit and version'
	@echo '    make test            Run tests on a compiled project.'
	@echo '    make push            Push tagged images to registry'
	@echo '    make clean           Clean the directory tree.'
	@echo

build:
	@echo "building ${BIN_NAME} ${VERSION}"
	@echo "GOPATH=${GOPATH}"
	go build -ldflags "${FLAGS}" -o bin/${BIN_NAME}

run:
	make build
	bin/${BIN_NAME}


build-alpine:
	@echo "building ${BIN_NAME} ${VERSION}"
	@echo "GOPATH=${GOPATH}"
	GOOS=linux GOARCH=amd64 go build -ldflags ' -w -s ${FLAGS} ' -o bin/${BIN_NAME}

build-image:
	@echo "building image ${BIN_NAME} ${VERSION} $(GIT_COMMIT)"
	docker build	--build-arg VERSION=${VERSION} \
	 				--build-arg GIT_COMMIT=$(GIT_COMMIT) \
	 				--build-arg GIT_COMMIT_SHORT=$(GIT_COMMIT_SHORT) \
	 				--build-arg GIT_DIRTY=$(GIT_DIRTY) \
	 				--build-arg BUILD_DATE=$(BUILD_DATE) \
	 				-t $(IMAGE_NAME):local .

tag:
	@echo "Tagging: latest ${VERSION} $(GIT_COMMIT)"
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):$(GIT_COMMIT)
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):${VERSION}
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):latest

tag-image:
	@echo "Tagging: latest ${VERSION} $(GIT_COMMIT)"
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):${VERSION}
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):latest

push-image: tag-image
	@echo "Pushing docker image to registry: latest ${VERSION} $(GIT_COMMIT)"
	docker push $(IMAGE_NAME):$(GIT_COMMIT)
	docker push $(IMAGE_NAME):${VERSION}
	docker push $(IMAGE_NAME):latest

clean:
	@test ! -e bin/${BIN_NAME} || rm bin/${BIN_NAME}

test:
	go test ./...

