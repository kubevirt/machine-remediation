package e2e

import (
	"flag"
	"testing"

	"github.com/golang/glog"
	bmov1 "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/kubernetes/scheme"

	mrv1 "kubevirt.io/machine-remediation-operator/pkg/apis/machineremediation/v1alpha1"
	testsutils "kubevirt.io/machine-remediation-operator/tests/utils"

	mapiv1 "sigs.k8s.io/cluster-api/pkg/apis/machine/v1beta1"
)

var (
	// ContainerPrefix contains the registry of images used for testing
	ContainerPrefix = "docker.io/kubevirt"
	// ContainerRegistry contains the tag of images used for testing
	ContainerTag = "latest"
)

func init() {
	testsutils.Init()
	flag.StringVar(&ContainerPrefix, "container-prefix", "docker.io/kubevirt", "Set the repository prefix for all images")
	flag.StringVar(&ContainerTag, "container-tag", "latest", "Set the image tag or digest to use")
	flag.Parse()

	if err := mrv1.AddToScheme(scheme.Scheme); err != nil {
		glog.Fatal(err)
	}

	if err := bmov1.SchemeBuilder.AddToScheme(scheme.Scheme); err != nil {
		glog.Fatal(err)
	}

	if err := mapiv1.AddToScheme(scheme.Scheme); err != nil {
		glog.Fatal(err)
	}

}

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Remediation Test Suite")
}
