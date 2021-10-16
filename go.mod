module github.com/waterme7on/openGauss-operator

go 1.16

require (
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.26.0
	golang.org/x/net v0.0.0-20210902165921-8d991716f632 // indirect
	golang.org/x/sys v0.0.0-20210903071746-97244b99971b // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20210831024726-fe130286e0e2 // indirect
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.21.2
	k8s.io/code-generator v0.0.0-20210313024825-5bc604e30f37
	k8s.io/klog/v2 v2.9.0
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20210313025757-51a1c5553d68
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20210313025227-57f2a0733447
	// k8s.io/client-go => k8s.io/client-go v0.0.0-20210313030403-f6ce18ae578c
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20210313024825-5bc604e30f37
)
