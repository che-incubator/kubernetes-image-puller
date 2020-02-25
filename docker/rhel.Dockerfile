# Copyright (c) 2020 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#
# Contributors:
#   Red Hat, Inc. - initial API and implementation
#

# NOTE: using registry.access.redhat.com/rhel8/go-toolset does not work (user is requested to use registry.redhat.io)
# NOTE: using registry.redhat.io/rhel8/go-toolset requires login, which complicates automation
# https://access.redhat.com/containers/?tab=tags#/registry.access.redhat.com/devtools/go-toolset-rhel7
FROM registry.access.redhat.com/devtools/go-toolset-rhel7:1.12.12-4 as builder

ARG BOOTSTRAP=true
ENV BOOTSTRAP=${BOOTSTRAP}

ENV PATH=/opt/rh/go-toolset-1.12/root/usr/bin:$PATH \
    GO111MODULE=on

USER root

WORKDIR /kubernetes-image-puller
COPY go.mod .
COPY go.sum .
# built in Brew, use tarball in lookaside cache; built locally, comment this out
# COPY resources.tgz /tmp/resources.tgz
# build locally, fetch mods
RUN if [[ ${BOOTSTRAP} != "false" ]]; then \
      go mod download; \
    elif [[ -f /tmp/resources.tgz ]]; then \
      tar xvf /tmp/resources.tgz -C /; \
      rm -f /tmp/resources.tgz; \
    fi

COPY . .

RUN adduser appuser && \
    make build 

# https://access.redhat.com/containers/?tab=tags#/registry.access.redhat.com/ubi8-minimal
FROM registry.access.redhat.com/ubi8-minimal:8.1-398
USER root
# CRW-528 copy actual cert
COPY --from=builder /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem /etc/pki/ca-trust/extracted/pem/
# CRW-528 copy symlink to the above cert
COPY --from=builder /etc/pki/tls/certs/ca-bundle.crt                  /etc/pki/tls/certs/
COPY --from=builder /etc/passwd /etc/passwd

USER appuser
COPY --from=builder /kubernetes-image-puller/bin/kubernetes-image-puller /
# NOTE: To use this container, need a configmap. See example at ./deploy/openshift/configmap.yaml
# See also https://github.com/che-incubator/kubernetes-image-puller#configuration
CMD ["/kubernetes-image-puller"]

# append Brew metadata here
