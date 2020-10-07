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
	"strconv"
	"strings"

	orgv1 "github.com/eclipse/che-operator/pkg/apis/org/v1"
	"github.com/eclipse/che-operator/pkg/util"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	KeycloakDeploymentName   = "keycloak"
	selectSslRequiredCommand = "OUT=$(psql keycloak -tAc \"SELECT 1 FROM REALM WHERE id = 'master'\"); " +
		"if [ $OUT -eq 1 ]; then psql keycloak -tAc \"SELECT ssl_required FROM REALM WHERE id = 'master'\"; fi"
	updateSslRequiredCommand = "psql keycloak -c \"update REALM set ssl_required='NONE' where id = 'master'\""
)

var (
	trustpass              = util.GeneratePasswd(12)
	keycloakCustomDiffOpts = cmp.Options{
		cmp.Comparer(func(x, y appsv1.Deployment) bool {
			return x.Annotations["che.self-signed-certificate.version"] == y.Annotations["che.self-signed-certificate.version"] &&
				x.Annotations["che.openshift-api-crt.version"] == y.Annotations["che.openshift-api-crt.version"] &&
				x.Annotations["che.keycloak-ssl-required-updated"] == y.Annotations["che.keycloak-ssl-required-updated"]
		}),
	}
	keycloakAdditionalDeploymentMerge = func(specDeployment *appsv1.Deployment, clusterDeployment *appsv1.Deployment) *appsv1.Deployment {
		clusterDeployment.Spec = specDeployment.Spec
		clusterDeployment.Annotations["che.self-signed-certificate.version"] = specDeployment.Annotations["che.self-signed-certificate.version"]
		clusterDeployment.Annotations["che.openshift-api-crt.version"] = specDeployment.Annotations["che.openshift-api-crt.version"]
		clusterDeployment.Annotations["che.keycloak-ssl-required-updated"] = specDeployment.Annotations["che.keycloak-ssl-required-updated"]
		return clusterDeployment
	}
)

func SyncKeycloakDeploymentToCluster(deployContext *DeployContext) DeploymentProvisioningStatus {
	clusterDeployment, err := getClusterDeployment(KeycloakDeploymentName, deployContext.CheCluster.Namespace, deployContext.ClusterAPI.Client)
	if err != nil {
		return DeploymentProvisioningStatus{
			ProvisioningStatus: ProvisioningStatus{Err: err},
		}
	}

	specDeployment, err := getSpecKeycloakDeployment(deployContext, clusterDeployment)
	if err != nil {
		return DeploymentProvisioningStatus{
			ProvisioningStatus: ProvisioningStatus{Err: err},
		}
	}

	return SyncDeploymentToCluster(deployContext, specDeployment, clusterDeployment, keycloakCustomDiffOpts, keycloakAdditionalDeploymentMerge)
}

func getSpecKeycloakDeployment(
	deployContext *DeployContext,
	clusterDeployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	labels := GetLabels(deployContext.CheCluster, KeycloakDeploymentName)

	keycloakEnv := []corev1.EnvVar{
		// {
			// todo Proxy?
		// 	Name:  "PROXY_ADDRESS_FORWARDING",
		// 	Value: "true",
		// },
		{
			Name: "HOME",
			Value: "/opt/jboss",
		},
		{
			Name: "KEYCLOAK_USER",
			Value: "admin",
		},
		{
			Name: "KEYCLOAK_PASSWORD",
			Value: "admin",
		},
		{
			Name: "DB_VENDOR",
			Value: "H2",
		},
		{
			Name: "JGROUPS_DISCOVERY_PROTOCOL",
			Value: "dns.DNS_PING",
		},
		{
			Name: "JGROUPS_DISCOVERY_PROPERTIES",
			// JGROUPS_DISCOVERY_PROPERTIES=dns_query=keycloak.che.svc.cluster.local
			Value: "dns_query=" + KeycloakDeploymentName + "." + deployContext.CheCluster.Namespace + ".svc.cluster.local",
		},
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      KeycloakDeploymentName,
			Namespace: deployContext.CheCluster.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            KeycloakDeploymentName,
							Image:           "quay.io/keycloak/keycloak:11.0.2",
							ImagePullPolicy: "Always",
							Ports: []corev1.ContainerPort{
								{
									// Name:          KeycloakDeploymentName,
									ContainerPort: 8080,
									Protocol:      "TCP",
								},
								{
									Name: "https",
									ContainerPort: 8443,
									Protocol: "TCP",
								},
							},
							// Resources: corev1.ResourceRequirements{
							// 	Requests: corev1.ResourceList{
							// 		corev1.ResourceMemory: resource.MustParse("512Mi"),
							// 	},
							// 	Limits: corev1.ResourceList{
							// 		corev1.ResourceMemory: resource.MustParse("2Gi"),
							// 	},
							// },
							LivenessProbe: &corev1.Probe{
								FailureThreshold: 3,
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/auth/realms/master",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: int32(8080),
										},
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 60,
							},
							ReadinessProbe: &corev1.Probe{
								FailureThreshold:    10,
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/auth/realms/master",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: int32(8080),
										},
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 30,
							},
							Env: keycloakEnv,
						},
					},
					SecurityContext: &corev1.PodSecurityContext{},
					// TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					RestartPolicy:                 "Always",
				},
			},
		},
	}

//   - apiVersion: v1
//     kind: DeploymentConfig
//     metadata:
//       labels:
//         application: '${APPLICATION_NAME}'
//       name: '${APPLICATION_NAME}'
//     spec:
//       replicas: 1
//       selector:
//         deploymentConfig: '${APPLICATION_NAME}'
//       strategy:
//         type: Recreate
//       template:
//         metadata:
//           labels:
//             application: '${APPLICATION_NAME}'
//             deploymentConfig: '${APPLICATION_NAME}'
//           name: '${APPLICATION_NAME}'
//         spec:
//           containers:
//             - env:
//                 - name: KEYCLOAK_USER
//                   value: '${KEYCLOAK_USER}'
//                 - name: KEYCLOAK_PASSWORD
//                   value: '${KEYCLOAK_PASSWORD}'
//                 - name: DB_VENDOR
//                   value: '${DB_VENDOR}'
//                 - name: JGROUPS_DISCOVERY_PROTOCOL
//                   value: dns.DNS_PING
//                 - name: JGROUPS_DISCOVERY_PROPERTIES
//                   value: 'dns_query=${APPLICATION_NAME}.${NAMESPACE}.svc.cluster.local'
//               image: quay.io/keycloak/keycloak:11.0.2
//               name: '${APPLICATION_NAME}'
//               ports:
//                 - containerPort: 8080
//                   protocol: TCP
//                 - containerPort: 8443
//                   name: https
//                   protocol: TCP
//               livenessProbe:
//                 failureThreshold: 3
//                 httpGet:
//                   path: /auth/realms/master
//                   port: 8080
//                   scheme: HTTP
//                 initialDelaySeconds: 60
//               readinessProbe:
//                 failureThreshold: 10
//                 httpGet:
//                   path: /auth/realms/master
//                   port: 8080
//                   scheme: HTTP
//                 initialDelaySeconds: 30
//               securityContext:
//                 privileged: false

	if !util.IsTestMode() {
		err := controllerutil.SetControllerReference(deployContext.CheCluster, deployment, deployContext.ClusterAPI.Scheme)
		if err != nil {
			return nil, err
		}
	}

	return deployment, nil
}

func getSecretResourceVersion(name string, namespace string, clusterAPI ClusterAPI) string {
	secret := &corev1.Secret{}
	err := clusterAPI.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, secret)
	if err != nil {
		if !errors.IsNotFound(err) {
			logrus.Errorf("Failed to get %s secret: %s", name, err)
		}
		return ""
	}
	return secret.ResourceVersion
}

func isSslRequiredUpdatedForMasterRealm(deployContext *DeployContext) bool {
	if deployContext.CheCluster.Spec.Database.ExternalDb {
		return false
	}

	if util.IsTestMode() {
		return false
	}

	clusterDeployment, _ := getClusterDeployment(KeycloakDeploymentName, deployContext.CheCluster.Namespace, deployContext.ClusterAPI.Client)
	if clusterDeployment == nil {
		return false
	}

	value, err := strconv.ParseBool(clusterDeployment.ObjectMeta.Annotations["che.keycloak-ssl-required-updated"])
	if err == nil && value {
		return true
	}

	dbValue, _ := getSslRequiredForMasterRealm(deployContext.CheCluster)
	return dbValue == "NONE"
}

func getSslRequiredForMasterRealm(checluster *orgv1.CheCluster) (string, error) {
	podName, err := util.K8sclient.GetDeploymentPod(PostgresDeploymentName, checluster.Namespace)
	if err != nil {
		return "", err
	}

	stdout, err := util.K8sclient.ExecIntoPod(podName, selectSslRequiredCommand, "", checluster.Namespace)
	return strings.TrimSpace(stdout), err
}

func updateSslRequiredForMasterRealm(checluster *orgv1.CheCluster) error {
	podName, err := util.K8sclient.GetDeploymentPod(PostgresDeploymentName, checluster.Namespace)
	if err != nil {
		return err
	}

	_, err = util.K8sclient.ExecIntoPod(podName, updateSslRequiredCommand, "Update ssl_required to NONE", checluster.Namespace)
	return err
}

func ProvisionKeycloakResources(deployContext *DeployContext) error {
	keycloakProvisionCommand := GetKeycloakProvisionCommand(deployContext.CheCluster)
	podToExec, err := util.K8sclient.GetDeploymentPod(KeycloakDeploymentName, deployContext.CheCluster.Namespace)
	if err != nil {
		logrus.Errorf("Failed to retrieve pod name. Further exec will fail")
	}

	_, err = util.K8sclient.ExecIntoPod(podToExec, keycloakProvisionCommand, "create realm, client and user", deployContext.CheCluster.Namespace)
	return err
}
