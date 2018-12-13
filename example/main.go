package main

import (
	"context"
	"fmt"
	"github.com/gofunct/grpc12factor"
	"github.com/prometheus/common/log"
)

func main() {
	ctx := context.TODO()
	run := grpc12factor.NewRuntime()
	run = grpc12factor.Compose(grpc12factor.NewRuntime())

	api := newDemoServer()
	RegisterDemoServiceServer(run.Server, api)
	log.Fatal(run.Serve(ctx))
}

// DemoServiceServer defines a Server.
type MockServer struct{}

func newDemoServer() *MockServer {
	return &MockServer{}
}

// SayHello implements a interface defined by protobuf.
func (s *MockServer) SayHello(ctx context.Context, request *HelloRequest) (*HelloResponse, error) {
	return &HelloResponse{Message: fmt.Sprintf("Hello %s", request.Name)}, nil
}
