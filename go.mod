module github.com/openshift/machine-remediation-operator

require (
	github.com/cynepco3hahue/machine-health-check-operator v0.0.0-20190625154545-f9bf53bd55ca
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/markbates/inflect v1.0.4 // indirect
	github.com/metal3-io/baremetal-operator v0.0.0-20190705194231-6d5a9e11b6d0
	github.com/openshift/cluster-api v0.0.0-20190503193905-cad0f8326cd2 // indirect
	github.com/prometheus/client_golang v1.0.0 // indirect
	github.com/spf13/pflag v1.0.3
	go.uber.org/zap v1.10.0 // indirect
	k8s.io/api v0.0.0-20190627205229-acea843d18eb
	k8s.io/apiextensions-apiserver v0.0.0-20190629010545-2d36bfd0ff86 // indirect
	k8s.io/apimachinery v0.0.0-20190629125103-05b5762916b3
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.0.0-20190627204931-86aa6a6a5cf3
	k8s.io/klog v0.3.3 // indirect
	k8s.io/utils v0.0.0-20190607212802-c55fbcfc754a
	sigs.k8s.io/cluster-api v0.1.4
	sigs.k8s.io/controller-runtime v0.1.12
	sigs.k8s.io/controller-tools v0.1.11
)

replace github.com/metal3-io/baremetal-operator => github.com/cynepco3hahue/baremetal-operator v0.0.0-20190703074131-22b01a873954

replace k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190311093542-50b561225d70

replace k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab

replace k8s.io/apiextensions-apiserver => github.com/openshift/kubernetes-apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed

replace k8s.io/apimachinery => github.com/openshift/kubernetes-apimachinery v0.0.0-20190313205120-d7deff9243b1

replace sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.2.0-beta.1.0.20190520212815-96b67f231945

replace sigs.k8s.io/cluster-api => github.com/openshift/cluster-api v0.0.0-20190619113136-046d74a3bd91
