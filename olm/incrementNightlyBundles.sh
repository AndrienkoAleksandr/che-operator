#!/bin/bash
#
# Copyright (c) 2012-2020 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#
# Contributors:
#   Red Hat, Inc. - initial API and implementation

set -e

BASE_DIR=$(cd "$(dirname "$0")"; pwd)
ROOT_PROJECT_DIR=$(dirname "${BASE_DIR}")

source ${BASE_DIR}/check-yq.sh

for platform in 'kubernetes' 'openshift'
do
  OPM_BUNDLE_DIR="${ROOT_PROJECT_DIR}/deploy/olm-catalog/che-operator/eclipse-che-preview-${platform}"
  OPM_BUNDLE_MANIFESTS_DIR="${OPM_BUNDLE_DIR}/manifests"
  CSV="${OPM_BUNDLE_MANIFESTS_DIR}/che-operator.clusterserviceversion.yaml"

  currentNightlyVersion=$(yq -r ".spec.version" "${CSV}")
  echo  "[INFO] current nightly ${platform} version: ${currentNightlyVersion}"
  #  echo "${currentNightlyVersion}"
  versionWithoutNightly="${currentNightlyVersion%.nightly}"
  # echo "${versionWithoutNightly}"
  version="${versionWithoutNightly%-*}"
  # echo "${version}"
  incrementPart="${versionWithoutNightly#*-}"
  # echo "${incrementPart}"

  newVersion="${version}-$((incrementPart+1)).nightly"

  echo "[INFO] Set up nightly ${platform} version: ${newVersion}"
  yq -ryY "(.spec.version) = \"${newVersion}\" | (.metadata.name) = \"eclipse-che-preview-${platform}.v${newVersion}\"" "${CSV}" > "${CSV}.old"
  mv "${CSV}.old" "${CSV}"
done