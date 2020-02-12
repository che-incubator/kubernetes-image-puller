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

REGISTRY='quay.io'
ORGANIZATION='eclipse'
IMAGE_NAME='kubernetes-image-puller'

# Source build variables
function set_env_vars() {
  if [ -e "jenkins-env.json" ]; then
        eval "$(./env-toolkit load -f jenkins-env.json \
            DEVSHIFT_TAG_LEN \
            QUAY_ECLIPSE_CHE_USERNAME \
            QUAY_ECLIPSE_CHE_PASSWORD \
            JENKINS_URL \
            GIT_BRANCH \
            GIT_COMMIT \
            BUILD_NUMBER \
            ghprbSourceBranch \
            ghprbActualCommit \
            BUILD_URL \
            ghprbPullId)"
  fi
}

function install_deps() {
  # Update machine, get required deps in place
  yum -y -d 1 update
  yum -y -d 1 install epel-release
  yum -y -d 1 install --enablerepo=epel docker make golang git
  systemctl start docker

  # Login to quay.io
  docker login -u ${QUAY_ECLIPSE_CHE_USERNAME} -p ${QUAY_ECLIPSE_CHE_PASSWORD} ${REGISTRY}

  setup_golang
}

# Perform necessary GOPATH setup to make project buildable
function setup_golang() {
  go version
  mkdir -p $HOME/go $HOME/go/src $HOME/go/bin $HOME/go/pkg
  export GOPATH=$HOME/go
  export PATH=${GOPATH}/bin:$PATH
  mkdir -p ${GOPATH}/src/github.com/che-incubator
  cp -r $HOME/payload ${GOPATH}/src/github.com/che-incubator/kubernetes-image-puller
  cd ${GOPATH}/src/github.com/che-incubator/kubernetes-image-puller
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
docker build -t ${LOCAL_IMAGE_NAME} -f ./docker/Dockerfile.centos .
tag_and_push ${REGISTRY}/${ORGANIZATION}/${IMAGE_NAME}:${TAG}
tag_and_push ${REGISTRY}/${ORGANIZATION}/${IMAGE_NAME}:latest
