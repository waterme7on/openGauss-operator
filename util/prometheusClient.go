package util

import (
	"context"
	"time"

	api "github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"k8s.io/klog/v2"
)

func try() {
	// 连接到Prometheus Client
	config := api.Config{
		Address: "http://10.77.50.201:30364",
	}
	client, err := api.NewClient(config)
	if err != nil {
		klog.Fatalf("Connect to prometheus error: %s", err)
		return
	}
	// 执行query
	queryClient := v1.NewAPI(client)
	ctx := context.TODO()
	value, _, err := queryClient.Query(ctx, "sum(rate(container_cpu_usage_seconds_total{pod=~\"gourd.*\"}[1m])) by (pod)", time.Now())
	if err != nil {
		klog.Fatalf("Query error: %s", err)
		return
	}
	// query返回的结果类类型 Model.Value类型
	// 包含四种子类型: https://github.com/prometheus/common/blob/main/model/value.go#L237
	// 输出其中一个看看
	exp := value.(model.Vector)[0]
	klog.Info(exp.Metric) // labels https://github.com/prometheus/common/blob/8d1c9f84e3f78cb628a20f8e7be531c508237848/model/metric.go#L32
	klog.Info(exp.Value)  // float64 https://github.com/prometheus/common/blob/main/model/value.go#L113
	// result
	// [root@worker201 util]# go run prometheusClient.go
	// I0901 07:43:49.870115   23954 prometheusClient.go:30] {pod="gourdstore-slave4-65866ffbb4-x25lx"}
	// I0901 07:43:49.870268   23954 prometheusClient.go:31] 0.03698425872791319
}
