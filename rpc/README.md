# openGauss grpc Server

## Support API

```go
// get openGauss master and worker replication stats
Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error)
// scale master and worker replicas
Scale(ctx context.Context, req *pb.ScaleRequest) (*pb.ScaleResponse, error)
```

## Develop

[protobuf documentation](https://developers.google.com/protocol-buffers/docs/gotutorial)

```
# generate protobuf code
protoc --go_out=. --go_opt=paths=source_relative     --go-grpc_out=. --go-grpc_opt=paths=source_relative protobuf/clients.proto 
```


## Usage

see `rpc/rpcClient_test.go func TestRpcClient` 

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
ok      github.com/waterme7on/openGauss-operator/rpc  0.056s
```