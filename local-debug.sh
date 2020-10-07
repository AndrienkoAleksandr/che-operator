#!/bin/bash
#
# Copyright (c) 2019 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#
# Contributors:
#   Red Hat, Inc. - initial API and implementation

set -e

if [ $# -ne 1 ]; then
    echo -e "Wrong number of parameters.\nUsage: ./loca-debug.sh <custom-resource-yaml>\n"
    exit 1
fi

command -v delv >/dev/null 2>&1 || { echo "operator-sdk is not installed. Aborting."; exit 1; }
command -v operator-sdk >/dev/null 2>&1 || { echo -e $RED"operator-sdk is not installed. Aborting."$NC; exit 1; }

CHE_NAMESPACE=test2


set +e
kubectl create namespace $CHE_NAMESPACE
set -e

kubectl apply -f deploy/crds/org_v1_che_crd.yaml
kubectl apply -f $1 -n $CHE_NAMESPACE
cp -rf templates/keycloak_provision /tmp/keycloak_provision
cp -rf templates/oauth_provision /tmp/oauth_provision

operator-sdk up local --namespace=${CHE_NAMESPACE} --enable-delve
