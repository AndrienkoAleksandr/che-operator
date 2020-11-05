#!/bin/bash

set -e

OPM_BUNDLE_MANIFESTS_DIR="deploy/olm-catalog/eclipse-che-preview-openshift/manifests"
platform=openshift
package="eclipse-che-preview-${platform}"
channel="nightly"
imageTool="podman"

tag="0.3"
export CATALOG_BUNDLE_IMAGE="${IMAGE_REGISTRY_HOST}/${IMAGE_REGISTRY_USER_NAME}/che_operator_bundle:${tag}"
export CATALOG_SOURCE_IMAGE="${IMAGE_REGISTRY_HOST}/${IMAGE_REGISTRY_USER_NAME}/testing_catalog:${tag}"

export buildTool="podman"


opm alpha bundle generate \
--directory "${OPM_BUNDLE_MANIFESTS_DIR}" \
--package "${package}" \
--channels "${channel}" \
--default "${channel}"

opm alpha bundle build \
-d "${OPM_BUNDLE_MANIFESTS_DIR}" \
--tag "${CATALOG_BUNDLE_IMAGE}" \
--package "${package}" \
--channels "${channel}" \
--default "${channel}" \
--image-builder "${imageTool}"

podman push "${CATALOG_BUNDLE_IMAGE}" --tls-verify=false

opm index add \
--bundles "${CATALOG_BUNDLE_IMAGE}" \
--tag "${CATALOG_SOURCE_IMAGE}" \
--pull-tool "${imageTool}" \
--build-tool "${imageTool}" \
--mode semver \
--binary-image=quay.io/operator-framework/upstream-opm-builder:v1.15.1 \
--skip-tls

podman push "${CATALOG_SOURCE_IMAGE}" --tls-verify=false



# Include second bundle
./olm/update-nightly-bundle.sh

tag="0.4"
export CATALOG_BUNDLE_IMAGE_SECOND="${IMAGE_REGISTRY_HOST}/${IMAGE_REGISTRY_USER_NAME}/che_operator_bundle:${tag}"

opm alpha bundle build \
-d "${OPM_BUNDLE_MANIFESTS_DIR}" \
--tag "${CATALOG_BUNDLE_IMAGE_SECOND}" \
--package "${package}" \
--channels "${channel}" \
--default "${channel}" \
--image-builder "${imageTool}"

podman push "${CATALOG_BUNDLE_IMAGE_SECOND}" --tls-verify=false

opm index add \
--bundles "${CATALOG_BUNDLE_IMAGE_SECOND}" \
--tag "${CATALOG_SOURCE_IMAGE}" \
--pull-tool "${imageTool}" \
--build-tool "${imageTool}" \
--mode semver \
--generate \
--skip-tls \
--binary-image=quay.io/operator-framework/upstream-opm-builder:v1.15.1 \
--from-index "${CATALOG_SOURCE_IMAGE}"

podman push "${CATALOG_SOURCE_IMAGE}" --tls-verify=false
