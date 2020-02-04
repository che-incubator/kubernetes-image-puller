#!/bin/bash
#
# Copyright (c) 2019 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0

set -u
set -e

LOCAL_IMAGE_NAME='kubernetes-image-puller'
REGISTRY='quay.io'
ORGANIZATION='openshiftio'
RHEL_IMAGE_NAME='rhel-kubernetes-image-puller'
CENTOS_IMAGE_NAME='kubernetes-image-puller'

# Source build variables
function set_env_vars() {
  if [ -e "jenkins-env" ]; then
    cat jenkins-env \
      | grep -E "(DEVSHIFT_TAG_LEN|QUAY_USERNAME|QUAY_PASSWORD|GIT_COMMIT)=" \
      | sed 's/^/export /g' \
      > ~/.jenkins-env
    source ~/.jenkins-env
  fi
}

function install_deps() {
  # Update machine, get required deps in place
  yum -y -d 1 update
  yum -y -d 1 install epel-release
  yum -y -d 1 install --enablerepo=epel docker make golang git
  systemctl start docker

  # Login to quay.io
  docker login -u ${QUAY_USERNAME} -p ${QUAY_PASSWORD} ${REGISTRY}

  setup_golang
}

# Perform necessary GOPATH setup to make project buildable
function setup_golang() {
  go version
  mkdir -p $HOME/go $HOME/go/src $HOME/go/bin $HOME/go/pkg
  export GOPATH=$HOME/go
  export PATH=${GOPATH}/bin:$PATH
  mkdir -p ${GOPATH}/src/github.com/redhat-developer
  cp -r $HOME/payload ${GOPATH}/src/github.com/redhat-developer/kubernetes-image-puller
  cd ${GOPATH}/src/github.com/redhat-developer/kubernetes-image-puller
}

# Simplify tagging and pushing
function tag_and_push() {
  local tag
  tag=$1
  docker tag ${LOCAL_IMAGE_NAME} $tag
  docker push $tag | cat
}

# Cleanup on exit
function cleanup() {
  make clean
  if [ -f "./jenkins-env" ]; then
    rm ~/.jenkins-env
  fi
}
trap cleanup EXIT

set_env_vars
install_deps

# Build main executable and docker image, push to quay.io
make build
TAG=$(echo $GIT_COMMIT | cut -c1-${DEVSHIFT_TAG_LEN})
if [[ ${TARGET:-"centos"} = 'rhel' ]]; then
  docker build -t ${LOCAL_IMAGE_NAME} -f ./docker/Dockerfile.rhel . | cat
  tag_and_push ${REGISTRY}/${ORGANIZATION}/${RHEL_IMAGE_NAME}:${TAG}
  tag_and_push ${REGISTRY}/${ORGANIZATION}/${RHEL_IMAGE_NAME}:latest
else
  docker build -t ${LOCAL_IMAGE_NAME} -f ./docker/Dockerfile.centos . | cat
  tag_and_push ${REGISTRY}/${ORGANIZATION}/${CENTOS_IMAGE_NAME}:${TAG}
  tag_and_push ${REGISTRY}/${ORGANIZATION}/${CENTOS_IMAGE_NAME}:latest
fi
