#!/bin/bash
#
# Copyright (c) 2012-2021 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#
# Contributors:
#   Red Hat, Inc. - initial API and implementation

if [ -z "${OPERATOR_REPO}" ]; then
  SCRIPT=$(readlink -f "$0")
  OPERATOR_REPO=$(dirname "$(dirname "$SCRIPT")");
fi

source ${OPERATOR_REPO}/olm/check-yq.sh

platform=$1
if [ "${platform}" == "" ]; then
  echo "Please specify platform ('openshift' or 'kubernetes') as the first argument."
  echo ""
  echo "testUpdate.sh <platform> [<channel>] [<namespace>]"
  exit 1
fi

channel=$2
if [ "${channel}" == "" ]; then
  channel="nightly"
fi

namespace=$3
if [ "${namespace}" == "" ]; then
  namespace="eclipse-che-preview-test"
fi

init() {
  IMAGE_REGISTRY_HOST=${IMAGE_REGISTRY_HOST:-quay.io}
  IMAGE_REGISTRY_USER_NAME=${IMAGE_REGISTRY_USER_NAME:-eclipse}
  source "${OPERATOR_REPO}/olm/olm.sh"

  OPM_BUNDLE_DIR=$(getBundlePath "${platform}" "${channel}")
  packageName=$(getPackageName "${platform}")
  CSV_FILE_PATH="${OPM_BUNDLE_DIR}/manifests/${packageName}.clusterserviceversion.yaml"

  # CATALOG_BUNDLE_IMAGE="${IMAGE_REGISTRY_HOST}/${IMAGE_REGISTRY_USER_NAME}/che_operator_bundle:0.0.1"
  # CATALOG_IMAGENAME="${IMAGE_REGISTRY_HOST}/${IMAGE_REGISTRY_USER_NAME}/testing_catalog:0.0.1"

  # echo "[INFO] Build bundle image... ${CATALOG_BUNDLE_IMAGE}"
  # buildBundleImage "${platform}" "${CATALOG_BUNDLE_IMAGE}" "${channel}" "docker"

  # echo "[INFO] Build catalog image... ${CATALOG_BUNDLE_IMAGE}"
  # buildCatalogImage "${CATALOG_IMAGENAME}" "${CATALOG_BUNDLE_IMAGE}" "docker" "false"

  # echo "[INFO]: Successfully created catalog source container image and enabled minikube ingress."
}

run() {
  createNamespace "${namespace}"

  installOperatorMarketPlace

  installCatalogSource "${platform}" "${namespace}" "${CATALOG_IMAGENAME}"

  getBundleListFromCatalogSource "${platform}" "${namespace}"
  getPreviousCSVInfo "${channel}"
  getLatestCSVInfo "${channel}"

  forcePullingOlmImages "${namespace}" "${PREVIOUS_CSV_BUNDLE_IMAGE}"
  forcePullingOlmImages "${namespace}" "${LATEST_CSV_BUNDLE_IMAGE}"

  subscribeToInstallation "${platform}" "${namespace}" "${channel}" "${PREVIOUS_CSV_NAME}"
  installPackage "${platform}" "${namespace}"
  echo -e "\u001b[32m Installation of the previous che-operator version: ${PREVIOUS_CSV_NAME} successfully completed \u001b[0m"
  applyCRCheCluster "${platform}" "${namespace}" "${CSV_FILE_PATH}"
  waitCheServerDeploy "${namespace}"

  installPackage "${platform}" "${namespace}"
  echo -e "\u001b[32m Installation of the latest che-operator version: ${LATEST_CSV_NAME} successfully completed \u001b[0m"
}

init
run
echo -e "\u001b[32m Done. \u001b[0m"
