BINARY_NAME=kubernetes-image-puller
DOCKERIMAGE_NAME=kubernetes-image-puller
DOCKERIMAGE_TAG=latest

all: build docker

.PHONY: build docker

build: test
	GOOS=linux go build -v -o ./bin/${BINARY_NAME} ./cmd/main.go

lint:
	golangci-lint run -v

test:
	go test -v ./cfg... ./pkg... ./utils...

docker:
	docker build -t ${DOCKERIMAGE_NAME}:${DOCKERIMAGE_TAG} -f ./docker/Dockerfile .

docker-dev: build
	docker build -t ${DOCKERIMAGE_NAME}:${DOCKERIMAGE_TAG} -f ./docker/dev.Dockerfile .

local-setup:
	oc process -f ./deploy/serviceaccount.yaml | oc apply -f -

local-deploy:
	oc apply -f ./deploy/configmap.yaml
	oc process -f ./deploy/app.yaml | oc apply -f -

clean:
	rm -rf ./bin
