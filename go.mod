module kubevirt.io/machine-remediation-operator

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/markbates/inflect v1.0.4 // indirect
	github.com/metal3-io/baremetal-operator v0.0.0-20190705194231-6d5a9e11b6d0
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/openshift/api v0.0.0-00010101000000-000000000000
	github.com/openshift/cluster-api v0.0.0-20190503193905-cad0f8326cd2 // indirect
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190822225318-a96b2b847d08
	github.com/prometheus/client_golang v1.0.0 // indirect
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.3.0
	golang.org/x/crypto v0.0.0-20190621222207-cc06ce4a13d4 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	google.golang.org/genproto v0.0.0-20181016170114-94acd270e44e // indirect
	k8s.io/api v0.0.0-20190717022910-653c86b0609b
	k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	k8s.io/apimachinery v0.0.0-20190717022731-0bb8574e0887
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.0.0-20190627204931-86aa6a6a5cf3
	k8s.io/klog v0.3.3 // indirect
	k8s.io/kube-openapi v0.0.0-20190603182131-db7b694dc208 // indirect
	k8s.io/utils v0.0.0-20190712204705-3dccf664f023
	sigs.k8s.io/cluster-api v0.1.4
	sigs.k8s.io/controller-runtime v0.1.12
	sigs.k8s.io/controller-tools v0.1.11
)

replace github.com/metal3-io/baremetal-operator => github.com/openshift/baremetal-operator v0.0.0-20190727010256-d905ab5eaddd

replace github.com/openshift/api => github.com/openshift/api v3.9.1-0.20190904155310-a25bb2adc83e+incompatible

replace github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20190721020503-a85ea6a6b3a5

replace k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190311093542-50b561225d70

replace k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190313205120-d7deff9243b1

replace k8s.io/client-go => k8s.io/client-go v11.0.0+incompatible

replace sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v0.0.0-20190724202554-0c1d754dd648

replace sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.2.0-beta.1.0.20190520212815-96b67f231945

replace sigs.k8s.io/cluster-api => github.com/openshift/cluster-api v0.0.0-20190619113136-046d74a3bd91
