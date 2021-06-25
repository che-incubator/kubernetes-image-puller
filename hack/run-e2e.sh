#!/bin/bash

function cleanup() {
  kind delete cluster --name k8s-image-puller-e2e
}

if [[ $(kind get clusters | grep k8s-image-puller-e2e | wc -l) == '0' ]]; then 
  kind create cluster --name k8s-image-puller-e2e 
fi
kubectl config set current-context kind-k8s-image-puller-e2e
export CACHING_INTERVAL_HOURS='1'
export IMAGES='che-theia=quay.io/eclipse/che-theia:next;che-plugin-registry=quay.io/eclipse/che-plugin-registry:next'
go test -count=1 -v ./e2e... 

if [[ $1 == "--rm" ]]; then
  cleanup
fi
