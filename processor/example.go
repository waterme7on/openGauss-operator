package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/prometheus/common/model"
	pb "github.com/waterme7on/openGauss-operator/rpc/protobuf"
	"github.com/waterme7on/openGauss-operator/util/prometheusUtil"
	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("server_addr", "localhost:17173", "The server address in the format of host:port")
)

func main() {
	// skeleton code
	// 连接prometheus Client
	// address := "http://10.77.50.201:30364"
	address := "http://10.77.50.201:31111"
	_, queryClient, err := prometheusUtil.GetPrometheusClient(address)
	if err != nil {
		log.Fatalf("Cannot connect to prometheus: %s, %s", address, err.Error())
	}
	// 连接grpc server
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
	// TODO:
	// 查询: util/prometheusUtil/prometheusQuerys.go
	result, err := prometheusUtil.QueryPodCpuUsagePercentage("test-opengauss", queryClient)
	if err != nil {
		log.Fatalf("Cannot query prometheus: %s, %s", address, err.Error())
	}
	// 返回map[string]string
	m := extractResult(&result)
	fmt.Println(m)
	// 输出如：map[{pod="prometheus-6d75d99cb9-lx8w2"}:4.93641914680743 {pod="prometheus-adapter-5b8db7955f-6zs2j"}:0 {pod="prometheus-adapter-5b8db7955f-ktp2k"}:3.571457910076159 {pod="prometheus-k8s-0"}:311.1957729587634 {pod="prometheus-operator-75d9b475d9-955fv"}:0.6592752119650527]
	// key: {pod="prometheus-6d75d99cb9-lx8w2"}
	// value: 4.93641914680743
	// 均为string
	for k, v := range m {
		fmt.Println(k)
		fmt.Println(v)
	}
	if percentage, _ := strconv.ParseFloat(m["{pod=\"prometheus-operator-75d9b475d9-955fv\"}"], 64); percentage > 0 {
		fmt.Println("test scale out")
		request := &pb.ScaleRequest{
			OpenGaussObjectKey: "test/test-opengauss",
			MasterReplication:  1,
			WorkerReplication:  2,
		}
		response, err := client.Scale(context.TODO(), request)
		if err != nil {
			log.Fatal(err)
		}
		log.Print(response)
	}
	if percentage, _ := strconv.ParseFloat(m["xxx"], 64); percentage > 50 {
		// 示例如果cpu利用率大于50，则发起Scale调用
		// 发起rpc调用
		request := &pb.ScaleRequest{
			OpenGaussObjectKey: "test/test-opengauss",
			MasterReplication:  1,
		}
		response, err := client.Scale(context.TODO(), request)
		if err != nil {
			log.Fatal(err)
		}
		log.Print(response)
	}
}

func extractResult(v *model.Value) (m map[string]string) {
	switch (*v).(type) {
	case model.Vector:
		vec, _ := (*v).(model.Vector)
		m = vectorToMap(&vec)
	default:
		break
	}
	return
}

func vectorToMap(v *model.Vector) (m map[string]string) {
	m = make(map[string]string)
	for i := range *v {
		m[(*v)[i].Metric.String()] = (*v)[i].Value.String()
	}
	return
}
