DOCKER_USERNAME ?= synthao
IMAGE_NAME ?= sso
TAG ?= latest

.PHONY: all gen

all: build push

gen:
	@protoc -I proto proto/sso/sso.proto --go_out=./gen/go/ --go_opt=paths=source_relative --go-grpc_out=./gen/go/ --go-grpc_opt=paths=source_relative

build:
	@echo "Building Docker image ${IMAGE}"
	@docker build -t ${IMAGE} .

push:
	@echo "Pushing Docker image ${IMAGE}"
	@docker push ${IMAGE}