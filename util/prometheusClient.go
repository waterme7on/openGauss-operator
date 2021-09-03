package util

import (
	"errors"

	api "github.com/prometheus/client_golang/api"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
)

// GetPrometheusClient returns prometheus apiConfig and apiClient
func GetPrometheusClient(address string) (*api.Config, *prometheus.API, error) {
	// 连接到Prometheus Client
	config := api.Config{
		Address: address,
	}
	client, err := api.NewClient(config)
	if err != nil {
		return nil, nil, errors.New("Connect to prometheus error")
	}
	// 执行query
	queryClient := prometheus.NewAPI(client)
	return &config, &queryClient, nil
}
