#!/bin/bash

opm alpha bundle generate \
--directory deploy/olm-catalog/eclipse-che-preview-openshift/manifests \
--package eclipse-che-preview-openshift \
--channels nightly \
--default nightly