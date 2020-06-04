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
package che

import (
	"context"

	orgv1 "github.com/eclipse/che-operator/pkg/apis/org/v1"
	"github.com/eclipse/che-operator/pkg/deploy"
	"github.com/eclipse/che-operator/pkg/util"
	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileChe) getProxyConfiguration(checluster *orgv1.CheCluster) (*deploy.Proxy, error) {
	proxy, err := deploy.ReadCheClusterProxyConfiguration(checluster)
	if err != nil {
		return nil, err
	}

	if util.IsOpenShift4 {
		clusterProxy := &configv1.Proxy{}
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: "cluster"}, clusterProxy); err != nil {
			return nil, err
		}

		// proxy configuration in CR overrides cluster proxy configuration
		if proxy.HttpProxy == "" && clusterProxy.Status.HTTPProxy != "" {
			proxy, err = deploy.ReadClusterWideProxyConfiguration(clusterProxy)
			if err != nil {
				return nil, err
			}

		}
	}

	return proxy, nil
}

func (r *ReconcileChe) putProxyCertIntoTrustStoreConfigMap(checluster *orgv1.CheCluster, proxy *deploy.Proxy, clusterAPI deploy.ClusterAPI) (bool, error) {
	if checluster.Spec.Server.ServerTrustStoreConfigMapName == "" {
		checluster.Spec.Server.ServerTrustStoreConfigMapName = deploy.DefaultCheServerCertConfigMap()
		if err := r.UpdateCheCRStatus(checluster, "Server Trust Store configmap", checluster.Spec.Server.ServerTrustStoreConfigMapName); err != nil {
			return false, err
		}
	}

	proxyCA, err := util.K8sclient.ReadOpenshiftConfigMap(proxy.TrustedCAMapName)
	if err != nil {
		return false, err
	}

	certConfigMap, err := deploy.SyncTrustStoreConfigMapToCluster(checluster, proxyCA.Data, clusterAPI)
	return certConfigMap != nil, err
}
