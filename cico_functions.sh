#!/bin/bash

function load_jenkins_vars() {
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
  REGISTRY="quay.io"
  # Update machine, get required deps in place

  /usr/sbin/setenforce 0 || true

  yum -y -d 1 update
  yum -y -d 1 install epel-release
  yum -y -d 1 install --enablerepo=epel docker make golang git
  systemctl start docker

  # Login to quay.io
  docker login -u "${QUAY_ECLIPSE_CHE_USERNAME}" -p "${QUAY_ECLIPSE_CHE_PASSWORD}" "${REGISTRY}"
  setup_golang
}

# Perform necessary GOPATH setup to make project buildable
function setup_golang() {
  go version
  mkdir -p "$HOME"/go "$HOME"/go/src "$HOME"/go/bin "$HOME"/go/pkg
  export GOPATH=$HOME/go
  export PATH=${GOPATH}/bin:$PATH
  mkdir -p "${GOPATH}"/src/github.com/che-incubator
  cp -r "$HOME"/payload "${GOPATH}"/src/github.com/che-incubator/kubernetes-image-puller
  cd "${GOPATH}"/src/github.com/che-incubator/kubernetes-image-puller || exit
}

function build() {
  LOCAL_IMAGE_NAME="kubernetes-image-puller"
  make build
  docker build -t ${LOCAL_IMAGE_NAME} -f ./docker/Dockerfile.centos .
}


function set_git_commit_tag() {
  GIT_COMMIT_TAG=$(echo "$GIT_COMMIT" | cut -c1-"${DEVSHIFT_TAG_LEN}")
  export GIT_COMMIT_TAG
}
# Simplify tagging and pushing
function tag_and_push_ci() {
  REGISTRY="quay.io"
  ORGANIZATION="eclipse"
  IMAGE="kubernetes-image-puller"
  LOCAL_IMAGE_NAME="kubernetes-image-puller"

  set_git_commit_tag
  docker tag ${LOCAL_IMAGE_NAME}  "${REGISTRY}/${ORGANIZATION}/${IMAGE}:${GIT_COMMIT_TAG}"
  docker push "${REGISTRY}/${ORGANIZATION}/${IMAGE}:${GIT_COMMIT_TAG}"
  docker tag ${LOCAL_IMAGE_NAME}  "${REGISTRY}/${ORGANIZATION}/${IMAGE}:latest"
  docker push "${REGISTRY}/${ORGANIZATION}/${IMAGE}:latest"
}

function tag_and_push_nightly() {
  REGISTRY="quay.io"
  ORGANIZATION="eclipse"
  IMAGE="kubernetes-image-puller"
  LOCAL_IMAGE_NAME="kubernetes-image-puller"
  docker tag ${LOCAL_IMAGE_NAME}  "${REGISTRY}/${ORGANIZATION}/${IMAGE}:nightly"
  docker push  "${REGISTRY}/${ORGANIZATION}/${IMAGE}:nightly"
}
