module github.com/waterme7on/openGauss-controller

go 1.16

require (
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/prometheus/common v0.26.0 // indirect
	k8s.io/api v0.0.0-20210313025757-51a1c5553d68
	k8s.io/apimachinery v0.0.0-20210313025227-57f2a0733447
	k8s.io/client-go v0.0.0-00010101000000-000000000000
	k8s.io/code-generator v0.0.0-20210313024825-5bc604e30f37
	k8s.io/klog/v2 v2.8.0
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20210313025757-51a1c5553d68
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20210313025227-57f2a0733447
	k8s.io/client-go => k8s.io/client-go v0.0.0-20210313030403-f6ce18ae578c
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20210313024825-5bc604e30f37
)
