module github.com/streamnsight/cert-manager-webhook-oci

go 1.18

// version of cert-manager used by KubeFlow 1.6
require github.com/jetstack/cert-manager v1.5.5

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/evanphx/json-patch v4.11.0+incompatible // indirect
	github.com/go-logr/logr v0.4.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/miekg/dns v1.1.34 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	golang.org/x/net v0.0.0-20210726213435-c6fcb2dbf985 // indirect
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914 // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d // indirect
	golang.org/x/text v0.3.5 // indirect
	golang.org/x/time v0.0.0-20210611083556-38a9dc6acbc6 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.21.3 // indirect
	// version of k8s API used by vert-manager
	k8s.io/apiextensions-apiserver v0.21.3 // indirect
	k8s.io/apimachinery v0.21.3 // indirect
	k8s.io/client-go v0.21.3 // indirect
	k8s.io/klog/v2 v2.8.0 // indirect
	k8s.io/kube-aggregator v0.21.3 // indirect
	k8s.io/kube-openapi v0.0.0-20210527164424-3c818078ee3d // indirect
	k8s.io/utils v0.0.0-20210802155522-efc7438f0176 // indirect
	sigs.k8s.io/controller-runtime v0.9.2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.2 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace (
	// Workaround to deal with https://github.com/kubernetes/klog/issues/253
	// Should be deleted when https://github.com/kubernetes/klog/pull/242 is merged and released
	github.com/go-logr/logr => github.com/go-logr/logr v0.4.0
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.1
	// per cert-manager go.mod
	golang.org/x/net => golang.org/x/net v0.0.0-20210224082022-3d97a244fca7

	// Per cert-manager go.mod
	google.golang.org/grpc => google.golang.org/grpc v1.29.1
	k8s.io/code-generator => github.com/kmodules/code-generator v0.21.1-rc.0.0.20210428003838-7eafae069eb0
	k8s.io/gengo => github.com/kmodules/gengo v0.0.0-20210428002657-a8850da697c2

	// See https://github.com/kubernetes/kubernetes/pull/99817
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20210305001622-591a79e4bda7
)
