// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	resourcemanager "github.com/gardener/gardener-resource-manager/pkg/manager"
	corev1 "k8s.io/api/core/v1"

	"github.com/gardener/gardener-extensions/controllers/networking-calico/pkg/charts"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
)

const (
	calicoSecretName       = "calico"
	calicoConfigSecretName = "calico-config"
	calicoRBACSecretName   = "calico-rbac"
	calicoPSPsSecretName   = "calico-psps"

	noCleanUp = "shoot.gardener.cloud/no-cleanup"
)

func withYamlSeparator(str string) string {
	return str + "\n" + "---" + "\n"
}

func withLocalObjectRefs(refs ...string) []corev1.LocalObjectReference {
	var localObjectRefs []corev1.LocalObjectReference
	for _, ref := range refs {
		localObjectRefs = append(localObjectRefs, corev1.LocalObjectReference{Name: ref})
	}
	return localObjectRefs
}

func Secrets(cl client.Client, calicoFiles *charts.CalicoFiles, namespace string) (*resourcemanager.Secrets, []corev1.LocalObjectReference) {
	calicoSecret := resourcemanager.NewSecret(cl).
		WithKeyValues(map[string][]byte{charts.CalicoKey: []byte(calicoFiles.Calico)}).
		WithNamespacedName(namespace, calicoSecretName)

	calicoConfigSecret := resourcemanager.NewSecret(cl).
		WithKeyValues(map[string][]byte{charts.CalicoConfigKey: []byte(calicoFiles.Config)}).
		WithNamespacedName(namespace, calicoConfigSecretName)

	calicoRBAC := resourcemanager.NewSecret(cl).
		WithKeyValues(map[string][]byte{charts.CalicoRBACKey: []byte(calicoFiles.RBAC)}).
		WithNamespacedName(namespace, calicoRBACSecretName)

	pspContents := withYamlSeparator(calicoFiles.CalicoTyphaClusterRoleBinding) +
		withYamlSeparator(calicoFiles.CalicoTyphaClusterRole) +
		withYamlSeparator(calicoFiles.CalicoClusterRole) +
		withYamlSeparator(calicoFiles.CalicoTyphaPSP) +
		withYamlSeparator(calicoFiles.CalicoClusterRoleBinding) +
		withYamlSeparator(calicoFiles.CalicoPSP) +
		withYamlSeparator(calicoFiles.CalicoKubeControllersClusterRole) +
		withYamlSeparator(calicoFiles.CalicoKubeControllersClusterRoleBinding) +
		calicoFiles.CalicoKubeControllersPSP

	calicoPSPs := resourcemanager.NewSecret(cl).
		WithKeyValues(map[string][]byte{"psps.yaml": []byte(pspContents)}).
		WithNamespacedName(namespace, calicoPSPsSecretName)

	return resourcemanager.NewSecrets(cl).WithSecretList([]resourcemanager.Secret{
		*calicoSecret,
		*calicoConfigSecret,
		*calicoRBAC,
		*calicoPSPs,
	}), withLocalObjectRefs(calicoSecretName, calicoConfigSecretName, calicoRBACSecretName, calicoPSPsSecretName)
}

// Reconcile implements Network.Actuator.
func (a *actuator) Reconcile(ctx context.Context, network *extensionsv1alpha1.Network) error {
	networkConfig, err := CalicoNetworkConfigFromNetworkResource(network)
	if err != nil {
		return err
	}

	calicoFiles, err := charts.RenderCalicoChart(a.chartRenderer, network, networkConfig)
	if err != nil {
		return err
	}

	secrets, secretRefs := Secrets(a.client, calicoFiles, network.Namespace)
	if err != nil {
		return err
	}

	err = secrets.Reconcile(ctx)
	if err != nil {
		return err
	}

	if err := resourcemanager.NewManagedResource(a.client).
		WithNamespacedName(network.Namespace, network.Name).
		WithSecretRefs(secretRefs).
		WithInjectedLabels(map[string]string{noCleanUp: "true"}).
		Reconcile(ctx); err != nil {
		return err
	}

	return a.updateProviderStatus(ctx, network, networkConfig)
}
