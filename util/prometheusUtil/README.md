
### 连接Prometheus

示例

```
address := "http://10.77.50.201:30364"  // 示例 asteria中的prometheus地址
_, queryClient, err := GetPrometheusClient(address)
if err != nil {
    t.Fatalf("Cannot connect to prometheus: %s, %s", address, err.Error())
}
```

### 执行查询

已有的查询从 `prometheusClient.go,prometheusQuerys.go` 中获得

示例
```
result, err = QueryPodMemoryUsage("prometheus", queryClient)  // 返回值为model.Value类型
if err != nil {
    t.Fatalf("Cannot query prometheus: %s, %s", address, err.Error())
}
```


### 测试

```
go test
```