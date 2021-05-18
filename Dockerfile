# Copyright (c) 2018-2021 Red Hat, Inc.
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
# NOTE: since updateBaseImages.sh does not support other registries than RHCC, update to RHEL8
# https://access.redhat.com/containers/?tab=tags#/registry.access.redhat.com/devtools/go-toolset-rhel7


# Build the manager binary
FROM registry.access.redhat.com/devtools/go-toolset-rhel7:1.14.12 as builder
ENV PATH=/opt/rh/go-toolset-1.14/root/usr/bin:${PATH} \
    GOPATH=/go/

ARG DEV_WORKSPACE_CONTROLLER_VERSION="main"
ARG DEV_WORKSPACE_CHE_OPERATOR_VERSION="main"

USER root

WORKDIR /che-operator
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY templates/ templates/
COPY pkg/ pkg/

# Build
RUN export ARCH="$(uname -m)" && if [[ ${ARCH} == "x86_64" ]]; then export ARCH="amd64"; elif [[ ${ARCH} == "aarch64" ]]; then export ARCH="arm64"; fi && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o che-operator main.go

# upstream, download devworkspace-operator templates for every build
# downstream, copy prefetched zip into /tmp
RUN curl -L https://api.github.com/repos/devfile/devworkspace-operator/zipball/${DEV_WORKSPACE_CONTROLLER_VERSION} > /tmp/devworkspace-operator.zip && \
    unzip /tmp/devworkspace-operator.zip */deploy/deployment/* -d /tmp && \
    mkdir -p /tmp/devworkspace-operator/templates/ && \
    mv /tmp/devfile-devworkspace-operator-*/deploy /tmp/devworkspace-operator/templates/

# upstream, download devworkspace-che-operator templates for every build
# downstream, copy prefetched zip into /tmp
RUN curl -L https://api.github.com/repos/che-incubator/devworkspace-che-operator/zipball/${DEV_WORKSPACE_CHE_OPERATOR_VERSION} > /tmp/devworkspace-che-operator.zip && \
    unzip /tmp/devworkspace-che-operator.zip */deploy/deployment/* -d /tmp && \
    mkdir -p /tmp/devworkspace-che-operator/templates/ && \
    mv /tmp/che-incubator-devworkspace-che-operator-*/deploy /tmp/devworkspace-che-operator/templates/

# https://access.redhat.com/containers/?tab=tags#/registry.access.redhat.com/ubi8-minimal
FROM registry.access.redhat.com/ubi8-minimal:8.3-298.1618432845

COPY --from=builder /che-operator/che-operator /manager
COPY --from=builder /che-operator/templates/keycloak-provision.sh /tmp/keycloak-provision.sh
COPY --from=builder /che-operator/templates/keycloak-update.sh /tmp/keycloak-update.sh
COPY --from=builder /che-operator/templates/oauth-provision.sh /tmp/oauth-provision.sh
COPY --from=builder /che-operator/templates/delete-identity-provider.sh /tmp/delete-identity-provider.sh
COPY --from=builder /che-operator/templates/create-github-identity-provider.sh /tmp/create-github-identity-provider.sh
COPY --from=builder /tmp/devworkspace-operator/templates/deploy /tmp/devworkspace-operator/templates
COPY --from=builder /tmp/devworkspace-che-operator/templates/deploy /tmp/devworkspace-che-operator/templates

WORKDIR /
USER 65532:65532

ENTRYPOINT ["/manager"]
