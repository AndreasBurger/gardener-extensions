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

package packet_test

import (
	. "github.com/gardener/gardener-extensions/controllers/provider-packet/pkg/packet"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Secret", func() {
	var secret *corev1.Secret

	BeforeEach(func() {
		secret = &corev1.Secret{}
	})

	Describe("#ReadCredentialsSecret", func() {
		It("should return an error because api token is missing", func() {
			credentials, err := ReadCredentialsSecret(secret)

			Expect(credentials).To(BeNil())
			Expect(err).To(HaveOccurred())
		})

		It("should return an error because project id is missing", func() {
			secret.Data = map[string][]byte{
				APIToken: []byte("foo"),
			}

			credentials, err := ReadCredentialsSecret(secret)

			Expect(credentials).To(BeNil())
			Expect(err).To(HaveOccurred())
		})

		It("should return the credentials structure", func() {
			var (
				apiToken  = []byte("foo")
				projectID = []byte("bar")
			)

			secret.Data = map[string][]byte{
				APIToken:  apiToken,
				ProjectID: projectID,
			}

			credentials, err := ReadCredentialsSecret(secret)

			Expect(credentials).To(Equal(&Credentials{
				APIToken:  apiToken,
				ProjectID: projectID,
			}))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
