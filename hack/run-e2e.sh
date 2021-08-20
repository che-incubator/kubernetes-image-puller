#!/bin/bash

export CACHING_INTERVAL_HOURS='1'
export DEPLOYMENT_NAME='kubernetes-image-puller'
export NAMESPACE='k8s-image-puller'
export DAEMONSET_NAME='kubernetes-image-puller'

function createClusterIfNeeded() {
  if [[ $(kind get clusters | grep k8s-image-puller-e2e | wc -l) == '0' ]]; then 
    kind create cluster --name k8s-image-puller-e2e 
  fi
}

function setContext() {
  kubectl config set current-context kind-k8s-image-puller-e2e
}

function setupCluster() {
  kubectl create namespace k8s-image-puller
  helm template -s templates/app.yaml ./deploy/helm | kubectl apply -n k8s-image-puller -f -
}

function cleanupCluster() {
  kubectl delete namespace k8s-image-puller --wait=true
}

function deleteCuster() {
  kind delete cluster --name k8s-image-puller-e2e
}

createClusterIfNeeded
setContext
setupCluster
go test -count=1 -v ./e2e...
cleanupCluster

if [[ $1 == "--rm" ]]; then
  deleteCuster
fi
