# openGauss grpc Server

## Support API

```go
// scale master and worker replicas
Scale(ctx context.Context, req *pb.ScaleRequest) (*pb.ScaleResponse, error)
```

## Testing

First run rpc server

```sh
go test -run TestRpcServer -kubeconfig=$HOME/.kube/config -timeout 30s
```

Then run rpcClient test
```sh
go test -run TestRpcClient
```

Then you will see outputs like
```
2021/09/04 00:07:44 success:true
PASS
ok      github.com/waterme7on/openGauss-controller/rpc  0.056s
```