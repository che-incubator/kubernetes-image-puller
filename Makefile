BINARY_NAME=kubernetes-image-puller
IMAGE_NAME=kubernetes-image-puller
IMAGE_TAG=next
CGO_ENABLED=1
PODMAN=podman

all: build docker

.PHONY: build docker

build: test
	CGO_ENABLED=${CGO_ENABLED} GOOS=linux go build -v -o ./bin/${BINARY_NAME} ./cmd/main.go
	CGO_ENABLED=${CGO_ENABLED} GOOS=linux go build -a -ldflags '-w -s' -a -installsuffix cgo -o ./bin/sleep ./sleep/sleep.go

lint:
	golangci-lint run -v

test:
	CGO_ENABLED=${CGO_ENABLED} go test -v ./cfg... ./pkg... ./sleep... ./utils...

docker:
	${PODMAN} build -t ${IMAGE_NAME}:${IMAGE_TAG} -f ./build/dockerfiles/Dockerfile .

docker-dev: build
	${PODMAN} build -t ${IMAGE_NAME}:${IMAGE_TAG} -f ./build/dockerfiles/dev.Dockerfile .

local-setup:
	oc process -f ./deploy/serviceaccount.yaml | oc apply -f -

local-deploy:
	oc apply -f ./deploy/configmap.yaml
	oc process -f ./deploy/app.yaml | oc apply -f -

clean:
	rm -rf ./bin
