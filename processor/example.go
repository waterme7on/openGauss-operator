package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/common/model"
	pb "github.com/waterme7on/openGauss-operator/rpc/protobuf"
	"github.com/waterme7on/openGauss-operator/util/prometheusUtil"
	"google.golang.org/grpc"
)

var (
	serverAddr    = flag.String("server_addr", "localhost:17173", "The server address in the format of host:port")
	address       = "http://10.77.50.201:31111"
	runtime       = time.Minute * 30  // 测试时间
	scaleInterval = time.Second * 300 // 弹性伸缩间隔
	isScale       = false             // 是否开启弹性伸缩
)

func main() {
	// skeleton code
	// 连接prometheus Client
	// address := "http://10.77.50.201:30364"
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
	// 获取集群的初始信息
	request := &pb.GetRequest{
		OpenGaussObjectKey: "test/a",
	}
	response, _ := client.Get(context.TODO(), request)
	scaleRequest := &pb.ScaleRequest{
		OpenGaussObjectKey: "test/a",
		MasterReplication:  response.MasterReplication,
		WorkerReplication:  response.WorkerReplication,
	}

	ctx, cancel := context.WithTimeout(context.TODO(), runtime)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		run(ctx, scaleRequest, client)
		defer wg.Done()
	}()
	log.Print("Wait for process done")
	wg.Wait()
	log.Print("Done")
}

func run(ctx context.Context, scaleRequest *pb.ScaleRequest, client pb.OpenGaussControllerClient) {
	_, queryClient, err := prometheusUtil.GetPrometheusClient(address)
	filePath := "slots.txt"
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	write := bufio.NewWriter(file)
	lastScaleTime := time.Now()
	if err != nil {
		log.Fatalf("Cannot connect to prometheus: %s, %s", address, err.Error())
	}
	for {
		time.Sleep(5 * time.Second)
		select {
		case <-ctx.Done():
			return
		default:
			result, err := prometheusUtil.QueryClusterCpuUsagePercentage("a", queryClient)
			if err != nil {
				log.Fatalf("Cannot query prometheus: %s, %s", address, err.Error())
			}
			// 返回map[string]string
			m := extractResult(&result)
			// 输出如：map[{pod="prometheus-6d75d99cb9-lx8w2"}:4.93641914680743 {pod="prometheus-adapter-5b8db7955f-6zs2j"}:0 {pod="prometheus-adapter-5b8db7955f-ktp2k"}:3.571457910076159 {pod="prometheus-k8s-0"}:311.1957729587634 {pod="prometheus-operator-75d9b475d9-955fv"}:0.6592752119650527]
			// key: {pod="prometheus-6d75d99cb9-lx8w2"}
			// value: 4.93641914680743
			// 均为string
			for _, v := range m {
				percentage, _ := strconv.ParseFloat(v, 64)
				fmt.Println("Cluster Cpu Usage", percentage)
				// ----------start-----------
				// 弹性伸缩相关代码：
				// 发起rpc调用
				if isScale {
					if percentage > 50 {
						scaleRequest.WorkerReplication += 1
						if time.Since(lastScaleTime) < scaleInterval {
							continue
						}
						response, err := client.Scale(context.TODO(), scaleRequest)
						if err != nil {
							log.Fatal(err)
						}
						log.Print(response)
					}
					if percentage < 5 {
						if scaleRequest.WorkerReplication <= 0 {
							continue
						}
						if time.Since(lastScaleTime) < scaleInterval {
							continue
						}
						scaleRequest.WorkerReplication = (scaleRequest.WorkerReplication - 1)
						response, err := client.Scale(context.TODO(), scaleRequest)
						if err != nil {
							log.Fatal(err)
						}
						log.Print(response)
					}
				}
				// ----------end-----------
			}
			result, err = prometheusUtil.QueryClusterNumber("a", queryClient)
			if err != nil {
				log.Fatalf("Cannot query prometheus: %s, %s", address, err.Error())
			}
			// 返回map[string]string
			m = extractResult(&result)
			for _, v := range m {
				write.WriteString(fmt.Sprintf("%v,%v\n", time.Now().Unix(), v))
			}
			write.Flush()
		}
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
