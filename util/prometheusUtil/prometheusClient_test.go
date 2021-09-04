package prometheusUtil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/common/model"
)

func TestPrometheusClient(t *testing.T) {
	address := "http://10.77.50.201:30364"
	_, queryClient, err := GetPrometheusClient(address)
	fmt.Println("Test util(TestPrometheusClient)")
	if err != nil {
		t.Fatalf("Cannot connect to prometheus: %s, %s", address, err.Error())
	}
	ctx := context.TODO()
	value, _, err := (*queryClient).Query(ctx, "sum(rate(container_cpu_usage_seconds_total{pod=~\"gourd.*\"}[1m])) by (pod)", time.Now())
	if err != nil {
		t.Fatalf(err.Error())
	}
	// query返回的结果类类型 Model.Value类型
	// 包含四种子类型: https://github.com/prometheus/common/blob/main/model/value.go#L237
	// 输出其中一个看看
	exp := value.(model.Vector)
	if exp.Len() > 0 {
		fmt.Printf("\tQuery Result: %s\n", exp[0])
	}
	fmt.Println("--Pass: TestPrometheusClient--")
}

func TestQuery(t *testing.T) {
	address := "http://10.77.50.201:30364"
	_, queryClient, err := GetPrometheusClient(address)
	fmt.Println("Test util(TestQuery)")
	if err != nil {
		t.Fatalf("Cannot connect to prometheus: %s, %s", address, err.Error())
	}
	result, err := QueryPodCpuUsage("prometheus", queryClient)
	if err != nil {
		t.Fatalf("Cannot query prometheus: %s, %s", address, err.Error())
	}
	fmt.Printf("Cpu Usage:\n%s\n", result)
	result, err = QueryPodCpuUsagePercentage("prometheus", queryClient)
	if err != nil {
		t.Fatalf("Cannot query prometheus: %s, %s", address, err.Error())
	}
	fmt.Printf("Cpu Usage Percentage:\n%s\n", result)
	result, err = QueryPodMemoryUsage("prometheus", queryClient)
	if err != nil {
		t.Fatalf("Cannot query prometheus: %s, %s", address, err.Error())
	}
	fmt.Printf("Memory Usage:\n%s\n", result)
	result, err = QueryPodMemoryUsagePercentage("prometheus", queryClient)
	if err != nil {
		t.Fatalf("Cannot query prometheus: %s, %s", address, err.Error())
	}
	fmt.Printf("Memory Usage Percentage:\n%s\n", result)
	fmt.Printf("--- Pass: TestQuery --\n")
}
