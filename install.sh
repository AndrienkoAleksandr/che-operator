#!/bin/bash

#  make bundle

make docker-build docker-push IMG="quay.io/aandriienko/che-operator:nightly"
# make bundle IMG="quay.io/aandriienko/che-operator-bundle:v1.0.2" platform="kubernetes"
make bundles
make bundle-build bundle-push platform="kubernetes"
# make catalog-build platform="kubernetes"
operator-sdk run bundle quay.io/aandriienko/che-operator-bundle:v1.0.2

# sleep 40

# oc apply -f config/samples/org.eclipse.che_v1_checluster.yaml


# operator-sdk cleanup eclipse-che-preview-kubernetes
