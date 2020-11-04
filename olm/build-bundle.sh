#!/bin/bash

readlink -f "$0"

# if [ -z "${OPERATOR_REPO}" ]; then
#   SCRIPT=$(readlink -f "$0")

#   OPERATOR_REPO=$(dirname "$(dirname "$SCRIPT")");
# fi
# echo "Operator repo path is ${OPERATOR_REPO}"

# OLM_DIR="${OPERATOR_REPO}/olm"
# export OPERATOR_REPO

source olm.sh openshift nightly che

installOPM

export CATALOG_BUNDLE_IMAGE="${IMAGE_REGISTRY_HOST}/${IMAGE_REGISTRY_USER_NAME}/che_operator_bundle:0.0.1"
export CATALOG_SOURCE_IMAGE="${IMAGE_REGISTRY_HOST}/${IMAGE_REGISTRY_USER_NAME}/testing_catalog:0.0.2"

export buildTool="podman"
export IMAGE_REGISTRY=quay.io
export IMAGE_REGISTRY_USERNAME=eclipse
export ECLIPSE_CATALOG_IMAGENAME="${IMAGE_REGISTRY}/${IMAGE_REGISTRY_USERNAME}/eclipse-che-${platform}-opm-catalog:preview"

export platform=openshift

echo "[INFO] Build bundle image... ${CATALOG_BUNDLE_IMAGE}"
buildBundleImage "${CATALOG_BUNDLE_IMAGE}" "${buildTool}"

echo "[INFO] Build catalog image... ${CATALOG_BUNDLE_IMAGE}"
buildCatalogImage "${CATALOG_SOURCE_IMAGE}" "${CATALOG_BUNDLE_IMAGE}" "${buildTool}" 
# "${ECLIPSE_CATALOG_IMAGENAME}"
