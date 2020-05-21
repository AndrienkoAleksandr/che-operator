//
// Copyright (c) 2012-2019 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//
package deploy

import (
	"context"

	orgv1 "github.com/eclipse/che-operator/pkg/apis/org/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

func SyncTrustStoreConfigMapToCluster(checluster *orgv1.CheCluster, data map[string]string, clusterAPI ClusterAPI) (*corev1.ConfigMap, error) {
	name := checluster.Spec.Server.ServerTrustStoreConfigMapName
	specConfigMap, err := GetSpecConfigMap(checluster, name, data, clusterAPI)
	if err != nil {
		return nil, err
	}

	clusterConfigMap, err := getClusterConfigMap(specConfigMap.Name, specConfigMap.Namespace, clusterAPI.Client)
	if err != nil {
		return nil, err
	}

	if clusterConfigMap == nil {
		logrus.Infof("Creating a new object: %s, name %s", specConfigMap.Kind, specConfigMap.Name)
		err := clusterAPI.Client.Create(context.TODO(), specConfigMap)
		return nil, err
	}

	// "ca-bundle.crt" is a key containing CA(s) in a PEM format
	if clusterConfigMap.Data["ca-bundle.crt"] == "" && specConfigMap.Data["ca-bundle.crt"] != "" {
		clusterConfigMap.Data["ca-bundle.crt"] = specConfigMap.Data["ca-bundle.crt"]
		logrus.Infof("Updating existed object: %s, name: %s", specConfigMap.Kind, specConfigMap.Name)
		err := clusterAPI.Client.Update(context.TODO(), clusterConfigMap)
		return nil, err
	}

	return clusterConfigMap, nil
}
