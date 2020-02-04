BINARY_NAME=kubernetes-image-puller
DOCKERIMAGE_NAME=kubernetes-image-puller
DOCKERIMAGE_TAG=latest

all: build docker

.PHONY: build docker

build:
	GOOS=linux go build -v -o ./bin/${BINARY_NAME} ./cmd/main.go

docker: build
	docker build -t ${DOCKERIMAGE_NAME}:${DOCKERIMAGE_TAG} -f ./docker/Dockerfile .

local-setup:
	oc process -f ./deploy/serviceaccount.yaml | oc apply -f -

local-deploy:
	oc apply -f ./deploy/configmap.yaml
	oc process -f ./deploy/app.yaml | oc apply -f -

clean:
	rm -rf ./bin
