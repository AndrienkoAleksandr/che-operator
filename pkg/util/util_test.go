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
	"os"
	"reflect"
	"testing"

	orgv1 "github.com/eclipse/che-operator/pkg/apis/org/v1"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/sirupsen/logrus"
)

const (
	expectedProxyURLWithUsernamePassword    = "https://user:password@myproxy.com:1234"
	expectedProxyURLWithoutUsernamePassword = "https://myproxy.com:1234"
	expectedNoProxy                         = "localhost,myhost.com"
)

// func TestGenerateProxyEnvs(t *testing.T) {

// 	proxyUrl, noProxy, _ := GenerateProxyEnvs(proxyHost, proxyPort, nonProxyHosts, proxyUser, proxyPassword, "", "")

// 	if !reflect.DeepEqual(proxyUrl, expectedProxyURLWithUsernamePassword) {
// 		t.Errorf("Test failed. Expected %s but got %s", expectedProxyURLWithUsernamePassword, proxyUrl)
// 	}

// 	if !reflect.DeepEqual(noProxy, expectedNoProxy) {
// 		t.Errorf("Test failed. Expected %s but got %s", expectedNoProxy, noProxy)

// 	}

// 	proxyUrl, _, _ = GenerateProxyEnvs(proxyHost, proxyPort, nonProxyHosts, "", proxyPassword, "", "")
// 	if !reflect.DeepEqual(proxyUrl, expectedProxyURLWithoutUsernamePassword) {
// 		t.Errorf("Test failed. Expected %s but got %s", expectedProxyURLWithoutUsernamePassword, proxyUrl)
// 	}

// }

func TestGenerateProxyJavaOpts(t *testing.T) {
	proxy := &Proxy{
		HttpProxy:    "https://user:password@myproxy.com:1234",
		HttpUser:     "user",
		HttpPassword: "password",
		HttpHost:     "myproxy.com",
		HttpPort:     "1234",

		HttpsProxy:    "https://user:password@myproxy.com:1234",
		HttpsUser:     "user",
		HttpsPassword: "password",
		HttpsHost:     "myproxy.com",
		HttpsPort:     "1234",

		NoProxy: "localhost,myhost.com",
	}

	if err := os.Setenv("KUBERNETES_SERVICE_HOST", "172.30.0.1"); err != nil {
		logrus.Errorf("Failed to set env %s", err)
	}

	javaOpts, _ := GenerateProxyJavaOpts(proxy)
	expectedJavaOpts := " -Dhttp.proxyHost=myproxy.com -Dhttp.proxyPort=1234 -Dhttps.proxyHost=myproxy.com " +
		"-Dhttps.proxyPort=1234 -Dhttp.nonProxyHosts='localhost|myhost.com' -Dhttp.proxyUser=user " +
		"-Dhttp.proxyPassword=password -Dhttps.proxyUser=user -Dhttps.proxyPassword=password"
	if !reflect.DeepEqual(javaOpts, expectedJavaOpts) {
		t.Errorf("Test failed. Expected '%s' but got '%s'", expectedJavaOpts, javaOpts)

	}

	proxy = &Proxy{
		HttpProxy: "https://user:password@myproxy.com:1234",
		HttpHost:  "myproxy.com",
		HttpPort:  "1234",

		HttpsProxy: "https://user:password@myproxy.com:1234",
		HttpsHost:  "myproxy.com",
		HttpsPort:  "1234",

		NoProxy: "localhost,myhost.com",
	}
	javaOpts, _ = GenerateProxyJavaOpts(proxy)
	expectedJavaOptsWithoutUsernamePassword := " -Dhttp.proxyHost=myproxy.com -Dhttp.proxyPort=1234 -Dhttps.proxyHost=myproxy.com " +
		"-Dhttps.proxyPort=1234 -Dhttp.nonProxyHosts='localhost|myhost.com'"

	if !reflect.DeepEqual(javaOpts, expectedJavaOptsWithoutUsernamePassword) {
		t.Errorf("Test failed. Expected '%s' but got '%s'", expectedJavaOptsWithoutUsernamePassword, javaOpts)

	}
}

func TestGeneratePasswd(t *testing.T) {
	chars := 12
	passwd := GeneratePasswd(chars)
	expectedCharsNumber := 12

	if !reflect.DeepEqual(len(passwd), expectedCharsNumber) {
		t.Errorf("Test failed. Expected %v chars, got %v chars", expectedCharsNumber, len(passwd))
	}

	passwd1 := GeneratePasswd(12)
	if reflect.DeepEqual(passwd, passwd1) {
		t.Errorf("Test failed. Passwords are identical, %s: %s", passwd, passwd1)
	}
}

func TestGetValue(t *testing.T) {
	key := "myvalue"
	defaultValue := "myDefaultValue"
	var1 := GetValue(key, defaultValue)
	var2 := GetValue("", defaultValue)

	if !reflect.DeepEqual(var1, key) {
		t.Errorf("Test failed. Expected '%s', but got '%s'", key, var1)
	}

	if !reflect.DeepEqual(var2, defaultValue) {
		t.Errorf("Test failed. Expected '%s', but got '%s'", var2, defaultValue)
	}
}

func TestReadCheClusterProxyConfiguration(t *testing.T) {
	checluster := &orgv1.CheCluster{
		Spec: orgv1.CheClusterSpec{
			Server: orgv1.CheClusterSpecServer{
				ProxyPassword: "password",
				ProxyUser:     "user",
				ProxyPort:     "1234",
				ProxyURL:      "https://myproxy.com",
				NonProxyHosts: "host1|host2",
			},
		},
	}
	expectedProxy := &Proxy{
		HttpProxy:    "https://user:password@myproxy.com:1234",
		HttpUser:     "user",
		HttpPassword: "password",
		HttpHost:     "myproxy.com",
		HttpPort:     "1234",

		HttpsProxy:    "https://user:password@myproxy.com:1234",
		HttpsUser:     "user",
		HttpsPassword: "password",
		HttpsHost:     "myproxy.com",
		HttpsPort:     "1234",

		NoProxy: "host1,host2",
	}

	actualProxy, _ := ReadCheClusterProxyConfiguration(checluster)

	if !reflect.DeepEqual(actualProxy, expectedProxy) {
		t.Errorf("Test failed. Expected '%v', but got '%v'", expectedProxy, actualProxy)
	}
}

func TestReadCheClusterProxyConfigurationNoUser(t *testing.T) {
	checluster := &orgv1.CheCluster{
		Spec: orgv1.CheClusterSpec{
			Server: orgv1.CheClusterSpecServer{
				ProxyPort:     "1234",
				ProxyURL:      "https://myproxy.com",
				NonProxyHosts: "host1|host2",
			},
		},
	}
	expectedProxy := &Proxy{
		HttpProxy: "https://myproxy.com:1234",
		HttpHost:  "myproxy.com",
		HttpPort:  "1234",

		HttpsProxy: "https://myproxy.com:1234",
		HttpsHost:  "myproxy.com",
		HttpsPort:  "1234",

		NoProxy: "host1,host2",
	}

	actualProxy, _ := ReadCheClusterProxyConfiguration(checluster)

	if !reflect.DeepEqual(actualProxy, expectedProxy) {
		t.Errorf("Test failed. Expected '%v', but got '%v'", expectedProxy, actualProxy)
	}
}

func TestReadCheClusterProxyConfigurationNoPort(t *testing.T) {
	checluster := &orgv1.CheCluster{
		Spec: orgv1.CheClusterSpec{
			Server: orgv1.CheClusterSpecServer{
				ProxyPassword: "password",
				ProxyUser:     "user",
				ProxyURL:      "https://myproxy.com",
				NonProxyHosts: "host1|host2",
			},
		},
	}
	expectedProxy := &Proxy{
		HttpProxy:    "https://user:password@myproxy.com",
		HttpUser:     "user",
		HttpPassword: "password",
		HttpHost:     "myproxy.com",

		HttpsProxy:    "https://user:password@myproxy.com",
		HttpsUser:     "user",
		HttpsPassword: "password",
		HttpsHost:     "myproxy.com",

		NoProxy: "host1,host2",
	}

	actualProxy, _ := ReadCheClusterProxyConfiguration(checluster)

	if !reflect.DeepEqual(actualProxy, expectedProxy) {
		t.Errorf("Test failed. Expected '%v', but got '%v'", expectedProxy, actualProxy)
	}
}

func TestReadClusterWideProxyConfiguration(t *testing.T) {
	clusterProxy := &configv1.Proxy{
		Status: configv1.ProxyStatus{
			HTTPProxy:  "http://user1:password1@myproxy1.com:1234",
			HTTPSProxy: "https://user2:password2@myproxy2.com:2345",
			NoProxy:    "host1,host2",
		},
	}

	expectedProxy := &Proxy{
		HttpProxy:    "http://user1:password1@myproxy1.com:1234",
		HttpUser:     "user1",
		HttpPassword: "password1",
		HttpHost:     "myproxy1.com",
		HttpPort:     "1234",

		HttpsProxy:    "https://user2:password2@myproxy2.com:2345",
		HttpsUser:     "user2",
		HttpsPassword: "password2",
		HttpsHost:     "myproxy2.com",
		HttpsPort:     "2345",

		NoProxy: "host1,host2",
	}

	actualProxy, _ := ReadClusterWideProxyConfiguration(clusterProxy)

	if !reflect.DeepEqual(actualProxy, expectedProxy) {
		t.Errorf("Test failed. Expected '%v', but got '%v'", expectedProxy, actualProxy)
	}
}

func TestReadClusterWideProxyConfigurationNoUser(t *testing.T) {
	clusterProxy := &configv1.Proxy{
		Status: configv1.ProxyStatus{
			HTTPProxy: "http://myproxy.com:1234",
			NoProxy:   "host1,host2",
		},
	}

	expectedProxy := &Proxy{
		HttpProxy: "http://myproxy.com:1234",
		HttpHost:  "myproxy.com",
		HttpPort:  "1234",

		HttpsProxy: "http://myproxy.com:1234",
		NoProxy:    "host1,host2",
		HttpsHost:  "myproxy.com",
		HttpsPort:  "1234",
	}

	actualProxy, _ := ReadClusterWideProxyConfiguration(clusterProxy)

	if !reflect.DeepEqual(actualProxy, expectedProxy) {
		t.Errorf("Test failed. Expected '%v', but got '%v'", expectedProxy, actualProxy)
	}
}

func TestReadClusterWideProxyConfigurationNoPort(t *testing.T) {
	clusterProxy := &configv1.Proxy{
		Status: configv1.ProxyStatus{
			HTTPProxy: "http://user:password@myproxy.com",
			NoProxy:   "host1,host2",
		},
	}

	expectedProxy := &Proxy{
		HttpProxy:    "http://user:password@myproxy.com",
		HttpUser:     "user",
		HttpPassword: "password",
		HttpHost:     "myproxy.com",

		HttpsProxy:    "http://user:password@myproxy.com",
		HttpsUser:     "user",
		HttpsPassword: "password",
		HttpsHost:     "myproxy.com",

		NoProxy: "host1,host2",
	}

	actualProxy, _ := ReadClusterWideProxyConfiguration(clusterProxy)

	if !reflect.DeepEqual(actualProxy, expectedProxy) {
		t.Errorf("Test failed. Expected '%v', but got '%v'", expectedProxy, actualProxy)
	}
}
