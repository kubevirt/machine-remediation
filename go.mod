module kubevirt.io/machine-remediation-operator

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0 // indirect
	github.com/metal3-io/baremetal-operator v0.0.0-20190705194231-6d5a9e11b6d0
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/openshift/api v3.9.1-0.20190517100836-d5b34b957e91+incompatible
	github.com/openshift/machine-api-operator v0.0.0-00010101000000-000000000000
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190822225318-a96b2b847d08
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.3.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apiextensions-apiserver v0.0.0-20190918201827-3de75813f604
	k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.0.0-20190912054826-cd179ad6a269
	k8s.io/utils v0.0.0-20190801114015-581e00157fb1
	sigs.k8s.io/cluster-api v0.1.4
	sigs.k8s.io/controller-runtime v0.3.1-0.20191016212439-2df793d02076
	sigs.k8s.io/controller-tools v0.2.2-0.20190919191502-76a25b63325a
)

replace (
	github.com/metal3-io/baremetal-operator => github.com/openshift/baremetal-operator v0.0.0-20190727010256-d905ab5eaddd
	github.com/openshift/api => github.com/openshift/api v3.9.1-0.20190904155310-a25bb2adc83e+incompatible
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20190923180330-3b6373338c9b
	github.com/openshift/machine-api-operator => github.com/openshift/machine-api-operator v0.2.1-0.20191025120018-fb3724fc7bdf
	k8s.io/api => k8s.io/api v0.0.0-20190918155943-95b840bb6a1f
	k8s.io/apiextensions-apiserver => github.com/openshift/kubernetes-apiextensions-apiserver v0.0.0-20190918161926-8f644eb6e783
	k8s.io/apimachinery => github.com/openshift/kubernetes-apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/client-go => github.com/openshift/kubernetes-client-go v0.0.0-20190918160344-1fbdaa4c8d90
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190912054826-cd179ad6a269
	sigs.k8s.io/cluster-api => github.com/openshift/cluster-api v0.0.0-20190619113136-046d74a3bd91
	sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.2.2-0.20190919191502-76a25b63325a
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v0.0.0-20190724202554-0c1d754dd648
)
