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

package charts_test

import (
	"fmt"
	"path/filepath"

	"github.com/gardener/gardener-extensions/controllers/networking-calico/pkg/calico"
	"github.com/gardener/gardener/pkg/chartrenderer"

	"k8s.io/helm/pkg/manifest"

	"github.com/golang/mock/gomock"

	calicov1alpha1 "github.com/gardener/gardener-extensions/controllers/networking-calico/pkg/apis/calico/v1alpha1"
	"github.com/gardener/gardener-extensions/controllers/networking-calico/pkg/charts"
	"github.com/gardener/gardener-extensions/controllers/networking-calico/pkg/imagevector"
	mockchartrenderer "github.com/gardener/gardener-extensions/pkg/mock/gardener/chartrenderer"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	YAML            = ".yaml"
	PSPPrefix       = "psp"
	CalicoKey       = "calico" + YAML
	CalicoConfigKey = "config" + YAML
	CalicoRBACKey   = "rbac" + YAML
)

var (
	CalicoClusterRoleKey        = filepath.Join(PSPPrefix, "calico-clusterrole"+YAML)
	CalicoClusterRoleBindingKey = filepath.Join(PSPPrefix, "calico-clusterrolebinding"+YAML)
	CalicoPSPKey                = filepath.Join(PSPPrefix, "calico-psp"+YAML)

	CalicoTyphaClusterRoleKey        = filepath.Join(PSPPrefix, "calico-typha-clusterrole"+YAML)
	CalicoTyphaClusterRoleBindingKey = filepath.Join(PSPPrefix, "calico-typha-clusterrolebinding"+YAML)
	CalicoTyphaPSPKey                = filepath.Join(PSPPrefix, "calico-typha-psp"+YAML)

	CalicoKubeControllersClusterRoleKey        = filepath.Join(PSPPrefix, "calico-kube-controllers-clusterrole"+YAML)
	CalicoKubeControllersClusterRoleBindingKey = filepath.Join(PSPPrefix, "calico-kube-controllers-clusterrolebinding"+YAML)
	CalicoKubeControllersPSPKey                = filepath.Join(PSPPrefix, "calico-kube-controllers-psp"+YAML)
)

var _ = Describe("Chart package test", func() {
	var (
		network       *extensionsv1alpha1.Network
		networkConfig *calicov1alpha1.NetworkConfig
		foorBar       = metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		}
	)

	BeforeEach(func() {
		network = &extensionsv1alpha1.Network{
			ObjectMeta: foorBar,
			Spec: extensionsv1alpha1.NetworkSpec{
				ServiceCIDR: "10.0.0.0/8",
				PodCIDR:     "192.168.1.0/24",
			},
		}
		networkConfig = &calicov1alpha1.NetworkConfig{}
	})

	Describe("#ComputeCalicoChartValues", func() {
		It("should correctly compute the calico chart values", func() {
			values := charts.ComputeCalicoChartValues(network, networkConfig)

			Expect(values).To(Equal(map[string]interface{}{
				"images": map[string]interface{}{
					"calico-cni":              imagevector.CalicoCNIImage(),
					"calico-typha":            imagevector.CalicoTyphaImage(),
					"calico-kube-controllers": imagevector.CalicoKubeControllersImage(),
					"calico-node":             imagevector.CalicoNodeImage(),
				},
				"global": map[string]string{
					"podCIDR": network.Spec.PodCIDR,
				},
			}))
		})
	})

	Describe("#RenderCalicoChart", func() {
		var (
			ctrl                = gomock.NewController(GinkgoT())
			mockChartRenderer   = mockchartrenderer.NewMockInterface(ctrl)
			testManifestContent = "test-content"
			expectedCalicoFiles = &charts.CalicoFiles{
				Calico: testManifestContent,
				Config: testManifestContent,
				RBAC:   testManifestContent,
				// PSP files
				CalicoClusterRole:        testManifestContent,
				CalicoClusterRoleBinding: testManifestContent,
				CalicoPSP:                testManifestContent,

				CalicoTyphaClusterRole:        testManifestContent,
				CalicoTyphaClusterRoleBinding: testManifestContent,
				CalicoTyphaPSP:                testManifestContent,

				CalicoKubeControllersClusterRole:        testManifestContent,
				CalicoKubeControllersClusterRoleBinding: testManifestContent,
				CalicoKubeControllersPSP:                testManifestContent,
			}

			mkManifest = func(name string) manifest.Manifest {
				return manifest.Manifest{Name: fmt.Sprintf("test/templates/%s", name), Content: testManifestContent}
			}
		)
		It("Render Calico Charts correctly", func() {
			mockChartRenderer.EXPECT().Render(calico.ChartPath, calico.ReleaseName, metav1.NamespaceSystem, gomock.Any()).Return(&chartrenderer.RenderedChart{
				ChartName: "test",
				Manifests: []manifest.Manifest{
					mkManifest(CalicoKey),
					mkManifest(CalicoConfigKey),
					mkManifest(CalicoRBACKey),
					mkManifest(CalicoClusterRoleKey),
					mkManifest(CalicoClusterRoleBindingKey),
					mkManifest(CalicoPSPKey),
					mkManifest(CalicoTyphaClusterRoleKey),
					mkManifest(CalicoTyphaClusterRoleBindingKey),
					mkManifest(CalicoTyphaPSPKey),
					mkManifest(CalicoKubeControllersClusterRoleKey),
					mkManifest(CalicoKubeControllersClusterRoleBindingKey),
					mkManifest(CalicoPSPKey),
				},
			}, nil)

			resultCalicoNetwork, err := charts.RenderCalicoChart(mockChartRenderer, network, networkConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(expectedCalicoFiles).To(Equal(resultCalicoNetwork))

		})
	})

})
