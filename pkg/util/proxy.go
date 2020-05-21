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

package util

import (
	"strings"

	"golang.org/x/net/http/httpproxy"

	"net/http"
	"net/url"

	orgv1 "github.com/eclipse/che-operator/pkg/apis/org/v1"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/sirupsen/logrus"
)

type Proxy struct {
	HttpProxy    string
	HttpUser     string
	HttpPassword string
	HttpHost     string
	HttpPort     string

	HttpsProxy    string
	HttpsUser     string
	HttpsPassword string
	HttpsHost     string
	HttpsPort     string

	NoProxy          string
	TrustedCAMapName string
}

func ReadClusterWideProxyConfiguration(clusterProxy *configv1.Proxy) (*Proxy, error) {
	proxy := &Proxy{}

	// Cluster components consume the status values to configure the proxy for their component.
	proxy.HttpProxy = clusterProxy.Status.HTTPProxy
	proxy.HttpsProxy = clusterProxy.Status.HTTPSProxy
	if len(proxy.HttpsProxy) == 0 {
		proxy.HttpsProxy = proxy.HttpProxy
	}
	proxy.NoProxy = clusterProxy.Status.NoProxy

	httpProxy, err := url.Parse(proxy.HttpProxy)
	if err != nil {
		return nil, err
	}
	proxy.HttpHost = httpProxy.Hostname()
	proxy.HttpPort = httpProxy.Port()
	proxy.HttpUser = httpProxy.User.Username()
	proxy.HttpPassword, _ = httpProxy.User.Password()

	httpsProxy, err := url.Parse(proxy.HttpsProxy)
	if err != nil {
		return nil, err
	}
	proxy.HttpsHost = httpsProxy.Hostname()
	proxy.HttpsPort = httpsProxy.Port()
	proxy.HttpsUser = httpsProxy.User.Username()
	proxy.HttpsPassword, _ = httpsProxy.User.Password()
	proxy.TrustedCAMapName = clusterProxy.Spec.TrustedCA.Name

	return proxy, nil
}

func ReadCheClusterProxyConfiguration(checluster *orgv1.CheCluster) (*Proxy, error) {
	proxyParts := strings.Split(checluster.Spec.Server.ProxyURL, "://")
	proxyProtocol := ""
	proxyHost := ""
	if len(proxyParts) == 1 {
		proxyProtocol = ""
		proxyHost = proxyParts[0]
	} else {
		proxyProtocol = proxyParts[0]
		proxyHost = proxyParts[1]
	}

	proxyURL := proxyHost
	if checluster.Spec.Server.ProxyPort != "" {
		proxyURL = proxyURL + ":" + checluster.Spec.Server.ProxyPort
	}

	proxyUser := checluster.Spec.Server.ProxyUser
	proxyPassword := checluster.Spec.Server.ProxyPassword
	proxySecret := checluster.Spec.Server.ProxySecret
	if len(proxySecret) > 0 {
		user, password, err := k8sclient.ReadSecret(proxySecret, checluster.Namespace)
		if err == nil {
			proxyUser = user
			proxyPassword = password
		} else {
			return nil, err
		}
	}

	if len(proxyUser) > 1 && len(proxyPassword) > 1 {
		proxyURL = proxyUser + ":" + proxyPassword + "@" + proxyURL
	}

	if proxyProtocol != "" {
		proxyURL = proxyProtocol + "://" + proxyURL
	}

	return &Proxy{
		HttpProxy:    proxyURL,
		HttpUser:     proxyUser,
		HttpHost:     proxyHost,
		HttpPort:     checluster.Spec.Server.ProxyPort,
		HttpPassword: proxyPassword,

		HttpsProxy:    proxyURL,
		HttpsUser:     proxyUser,
		HttpsHost:     proxyHost,
		HttpsPort:     checluster.Spec.Server.ProxyPort,
		HttpsPassword: proxyPassword,

		NoProxy: strings.Replace(checluster.Spec.Server.NonProxyHosts, "|", ",", -1),
	}, nil
}

func GenerateProxyJavaOpts(proxy *Proxy) (javaOpts string, err error) {
	noProxy := strings.Replace(proxy.NoProxy, ",", "|", -1)
	proxyUserPassword := ""
	if len(proxy.HttpUser) > 1 && len(proxy.HttpPassword) > 1 {
		proxyUserPassword = " -Dhttp.proxyUser=" + proxy.HttpUser + " -Dhttp.proxyPassword=" + proxy.HttpPassword +
			" -Dhttps.proxyUser=" + proxy.HttpsUser + " -Dhttps.proxyPassword=" + proxy.HttpsPassword
	}

	javaOpts =
		" -Dhttp.proxyHost=" + proxy.HttpHost + " -Dhttp.proxyPort=" + proxy.HttpPort +
			" -Dhttps.proxyHost=" + proxy.HttpsHost + " -Dhttps.proxyPort=" + proxy.HttpsPort +
			" -Dhttp.nonProxyHosts='" + noProxy + "'" + proxyUserPassword
	return javaOpts, nil
}

func ConfigureProxy(instance *orgv1.CheCluster, transport *http.Transport, proxy *Proxy) {
	config := httpproxy.Config{
		HTTPProxy:  proxy.HttpProxy,
		HTTPSProxy: proxy.HttpsProxy,
		NoProxy:    proxy.NoProxy,
	}
	proxyFunc := config.ProxyFunc()
	transport.Proxy = func(r *http.Request) (*url.URL, error) {
		theProxyUrl, err := proxyFunc(r.URL)
		if err != nil {
			logrus.Warnf("Error when trying to get the proxy to access TLS endpoint URL: %s - %s", r.URL, err)
		}
		logrus.Infof("Using proxy: %s to access TLS endpoint URL: %s", theProxyUrl, r.URL)
		return theProxyUrl, err
	}
}
