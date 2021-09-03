package util

import (
	"context"
	"errors"
	"fmt"
	"time"

	api "github.com/prometheus/client_golang/api"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// GetPrometheusClient returns prometheus apiConfig and apiClient
func GetPrometheusClient(address string) (*api.Config, *prometheus.API, error) {
	// 连接到Prometheus Client
	config := api.Config{
		Address: address,
	}
	client, err := api.NewClient(config)
	if err != nil {
		return nil, nil, errors.New("connect to prometheus error")
	}
	// 执行query
	queryClient := prometheus.NewAPI(client)
	return &config, &queryClient, nil
}

func QueryPodCpuUsage(podPrefix string, client *prometheus.API) (model.Value, error) {
	value, _, err := (*client).Query(context.TODO(), fmt.Sprintf(PodCpuUsage, podPrefix), time.Now())
	return value, err
}

func QueryPodCpuUsagePercentage(podPrefix string, client *prometheus.API) (model.Value, error) {
	value, _, err := (*client).Query(context.TODO(), fmt.Sprintf(PodCpuUsagePercentage, podPrefix), time.Now())
	return value, err
}

func QueryPodMemoryUsage(podPrefix string, client *prometheus.API) (model.Value, error) {
	value, _, err := (*client).Query(context.TODO(), fmt.Sprintf(PodMemoryUsage, podPrefix), time.Now())
	return value, err
}

func QueryPodMemoryUsagePercentage(podPrefix string, client *prometheus.API) (model.Value, error) {
	value, _, err := (*client).Query(context.TODO(), fmt.Sprintf(PodMemoryUsagePercentage, podPrefix), time.Now())
	return value, err
}
