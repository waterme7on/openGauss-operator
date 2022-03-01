module github.com/waterme7on/openGauss-operator

go 1.16

require (
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.26.0
	// github.com/waterme7on/openGauss-operator v0.0.0-20210929060813-0fc8f815caca
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.21.2
	k8s.io/code-generator v0.0.0-20210313024825-5bc604e30f37
	k8s.io/klog v1.0.0 // indirect
	k8s.io/klog/v2 v2.9.0
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20210313025757-51a1c5553d68
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20210313025227-57f2a0733447
	// k8s.io/client-go => k8s.io/client-go v0.0.0-20210313030403-f6ce18ae578c
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20210313024825-5bc604e30f37
)
