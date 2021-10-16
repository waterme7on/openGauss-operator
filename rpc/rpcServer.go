package rpc

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"errors"

	clientset "github.com/waterme7on/openGauss-operator/pkg/generated/clientset/versioned"
	pb "github.com/waterme7on/openGauss-operator/rpc/protobuf"
	"github.com/waterme7on/openGauss-operator/util"
	grpc "google.golang.org/grpc"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

var (
	port = flag.Int("port", 17173, "The server port")
)

type openGaussRpcServer struct {
	pb.UnimplementedOpenGaussControllerServer
	client clientset.Interface
	// mu     sync.Mutex // protects routeNotes
}

func (s *openGaussRpcServer) Scale(ctx context.Context, req *pb.ScaleRequest) (*pb.ScaleResponse, error) {
	klog.Infof("Rpc server receive request: %s", req.String())
	// Convert the namespace/name into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(req.OpenGaussObjectKey)
	response := &pb.ScaleResponse{Success: false}
	if err != nil {
		return response, errors.New("object key error, should be like namespace/name")
	}
	og, err := s.client.ControllerV1().OpenGausses(namespace).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return response, errors.New("error getting opengauss object")
	}
	og.Spec.OpenGauss.Master.Replicas = util.Int32Ptr(req.MasterReplication)
	og.Spec.OpenGauss.Worker.Replicas = util.Int32Ptr(req.WorkerReplication)
	_, err = s.client.ControllerV1().OpenGausses(namespace).Update(ctx, og, v1.UpdateOptions{})
	if err != nil {
		return response, errors.New("error during updata opengauss spec")
	}
	response.Success = true
	return response, nil
}

func (s *openGaussRpcServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	klog.Infof("Rpc server receive request: %s", req.String())
	// Convert the namespace/name into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(req.OpenGaussObjectKey)
	response := &pb.GetResponse{Success: false}
	if err != nil {
		return response, errors.New("object key error, should be like namespace/name")
	}
	og, err := s.client.ControllerV1().OpenGausses(namespace).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return response, errors.New("error getting opengauss object")
	}
	response.MasterReplication = (*og.Spec.OpenGauss.Master.Replicas)
	response.WorkerReplication = (*og.Spec.OpenGauss.Worker.Replicas)
	response.Success = true
	return response, nil
}

func NewRpcServer(openGaussClientset clientset.Interface) *openGaussRpcServer {
	s := &openGaussRpcServer{client: openGaussClientset}
	return s
}

// Run start Rpc server and listen to port
func (s *openGaussRpcServer) Run() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	klog.Infof("rpc server listening at localhost:%d", *port)
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterOpenGaussControllerServer(grpcServer, s)
	grpcServer.Serve(lis)
}
