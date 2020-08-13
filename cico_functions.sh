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

  yum install -d1 -y yum-utils device-mapper-persistent-data lvm2
  yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
  yum install -d1 -y docker-ce \
    git \
    make
  systemctl start docker

  # Login to quay.io
  docker login -u "${QUAY_ECLIPSE_CHE_USERNAME}" -p "${QUAY_ECLIPSE_CHE_PASSWORD}" "${REGISTRY}"
}

function build() {
  LOCAL_IMAGE_NAME="kubernetes-image-puller"
  docker build -t ${LOCAL_IMAGE_NAME} -f ./docker/Dockerfile .
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
