/*
 * This file is part of the KubeVirt Redfish project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2025 KubeVirt Redfish project and its authors.
 *
 */

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	redfishv1alpha1 "github.com/v1k0d3n/kubevirt-redfish-operator/api/v1alpha1"
)

var _ = Describe("RedfishServer Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		redfishserver := &redfishv1alpha1.RedfishServer{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind RedfishServer")
			err := k8sClient.Get(ctx, typeNamespacedName, redfishserver)
			if err != nil && errors.IsNotFound(err) {
				resource := &redfishv1alpha1.RedfishServer{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "redfish.kubevirt.io/v1alpha1",
						Kind:       "RedfishServer",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: redfishv1alpha1.RedfishServerSpec{
						Version:  "0.2.1",
						Replicas: 1,
						Chassis: []redfishv1alpha1.RedfishChassisSpec{
							{
								Name:        "test-chassis",
								Namespace:   "default",
								Description: "Test chassis for unit testing",
							},
						},
						Authentication: redfishv1alpha1.RedfishAuthSpec{
							Users: []redfishv1alpha1.RedfishUserSpec{
								{
									Username:       "test-user",
									PasswordSecret: "test-secret",
									Chassis:        []string{"test-chassis"},
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &redfishv1alpha1.RedfishServer{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance RedfishServer")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &RedfishServerReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify that the RedfishServer was created successfully
			By("Checking that the RedfishServer was created")
			createdRedfishServer := &redfishv1alpha1.RedfishServer{}
			err = k8sClient.Get(ctx, typeNamespacedName, createdRedfishServer)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdRedfishServer.Spec.Version).To(Equal("0.2.1"))
			Expect(createdRedfishServer.Spec.Replicas).To(Equal(int32(1)))
		})
	})
})
