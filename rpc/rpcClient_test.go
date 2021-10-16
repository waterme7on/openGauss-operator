package rpc

import (
	"context"
	"flag"
	"log"
	"testing"

	pb "github.com/waterme7on/openGauss-operator/rpc/protobuf"
	grpc "google.golang.org/grpc"
)

var (
	serverAddr = flag.String("server_addr", "localhost:17173", "The server address in the format of host:port")
)

func TestRpcClient(t *testing.T) {
	// 配置参数
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	// 建立连接
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	// 建立客户端
	client := pb.NewOpenGaussControllerClient(conn)
	// 发起rpc调用
	request := &pb.ScaleRequest{
		OpenGaussObjectKey: "default/test-opengauss",
		MasterReplication:  1,
	}
	response, err := client.Scale(context.TODO(), request)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(response)
}
