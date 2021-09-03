package util

const (
	PodCpuUsage              = "sum(rate(container_cpu_usage_seconds_total{pod=~\"%s.*\"}[1m])) by (pod)"
	PodCpuUsagePercentage    = "sum(rate(container_cpu_usage_seconds_total{pod=~\"%s.*\"}[1m])) by (pod) / (sum(container_spec_cpu_quota/100000) by (pod)) * 100"
	PodMemoryUsage           = "sum(container_memory_rss{pod=~\"%s.*\"}) by(pod)" // /1024/1024/1024 = GiB
	PodMemoryUsagePercentage = "100 * sum(container_memory_rss{pod=~\"%s.*\"}) by(pod) / sum(container_spec_memory_limit_bytes) by(pod)"
)
