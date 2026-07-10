#!/bin/bash
#
# Copyright (c) 2026 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#
# Release script for kubernetes-image-puller.
# Replaces :next image tags with the release version, commits, tags, and pushes.
#
# Usage: ./make-release.sh <version>
#   e.g. ./make-release.sh 7.99.0

set -e

VERSION=$1
if [ -z "${VERSION}" ]; then
  echo "Usage: $0 <version>"
  echo "  e.g. $0 7.99.0"
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "Releasing version ${VERSION}"

# Makefile: IMAGE_TAG=next -> IMAGE_TAG=<version>
sed -i "s/IMAGE_TAG=next/IMAGE_TAG=${VERSION}/" "${SCRIPT_DIR}/Makefile"

# Go default image constant
sed -i "s|kubernetes-image-puller:next|kubernetes-image-puller:${VERSION}|" "${SCRIPT_DIR}/cfg/envvars.go"

# Helm chart values
sed -i "s/  tag: next/  tag: ${VERSION}/" "${SCRIPT_DIR}/deploy/helm/values.yaml"

# OpenShift app template
sed -i "s/  value: next/  value: ${VERSION}/" "${SCRIPT_DIR}/deploy/openshift/app.yaml"

# OpenShift configmap
sed -i "s|kubernetes-image-puller:next|kubernetes-image-puller:${VERSION}|" "${SCRIPT_DIR}/deploy/openshift/configmap.yaml"

echo "Updated files:"
git diff --name-only

git add -A
git commit -m "chore: Release ${VERSION}"
git tag "v${VERSION}"
git push origin "v${VERSION}"

echo "Done. Tag v${VERSION} pushed — the release-build workflow will build and push the image."
