/*
Copyright 2019 The OpenEBS Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package localpv

import (
	"flag"

	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/maya/tests"
	"github.com/openebs/maya/tests/artifacts"

	ns "github.com/openebs/maya/pkg/kubernetes/namespace/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	kubeConfigPath string
	namespace      = "localpv-ns"
	namespaceObj   *corev1.Namespace
	pvcObj         *corev1.PersistentVolumeClaim
	pvcName        = "pvc-hp"
	scName         = "openebs-hostpath"
	accessModes    = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	capacity       = "2Gi"
	err            error
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test application deployment")
}

func init() {
	flag.StringVar(&kubeConfigPath, "kubeconfig", "", "path to kubeconfig to invoke kubernetes API calls")
}

var ops *tests.Operations

var _ = BeforeSuite(func() {

	ops = tests.NewOperations(tests.WithKubeConfigPath(kubeConfigPath))

	By("waiting for maya-apiserver pod to come into running state")
	podCount := ops.GetPodRunningCountEventually(
		string(artifacts.OpenebsNamespace),
		string(artifacts.MayaAPIServerLabelSelector),
		1,
	)
	Expect(podCount).To(Equal(1))

	By("waiting for openebs-provisioner pod to come into running state")
	podCount = ops.GetPodRunningCountEventually(
		string(artifacts.OpenebsNamespace),
		string(artifacts.OpenEBSLocalPVProvisionerLabelSelector),
		1,
	)
	Expect(podCount).To(Equal(1))

	By("building a namespace")
	namespaceObj, err = ns.NewBuilder().
		WithName(namespace).
		APIObject()
	Expect(err).ShouldNot(HaveOccurred(), "while building namespace {%s}", namespace)

	By("creating above namespace")
	_, err = ops.NSClient.Create(namespaceObj)
	Expect(err).To(BeNil(), "while creating namespace {%s}", namespaceObj.Name)

	By("building a pvc")
	pvcObj, err = pvc.NewBuilder().
		WithName(pvcName).
		WithNamespace(namespace).
		WithStorageClass(scName).
		WithAccessModes(accessModes).
		WithCapacity(capacity).Build()
	Expect(err).ShouldNot(
		HaveOccurred(),
		"while building pvc {%s} in namespace {%s}",
		pvcName,
		namespace,
	)

	By("creating above pvc")
	_, err = ops.PVCClient.WithNamespace(namespace).Create(pvcObj)
	Expect(err).To(
		BeNil(),
		"while creating pvc {%s} in namespace {%s}",
		pvcName,
		namespace,
	)

})

var _ = AfterSuite(func() {

	By("deleting above pvc")
	err = ops.PVCClient.Delete(pvcName, &metav1.DeleteOptions{})
	Expect(err).To(
		BeNil(),
		"while deleting pvc {%s} in namespace {%s}",
		pvcName,
		namespace,
	)

	By("deleting namespace")
	err = ops.NSClient.Delete(namespace, &metav1.DeleteOptions{})
	Expect(err).To(BeNil(), "while deleting namespace {%s}", namespaceObj.Name)

})
