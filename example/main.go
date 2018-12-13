package main

import (
	"context"
	"fmt"
	"github.com/gofunct/runtime"
	"github.com/prometheus/common/log"
)

func main() {
	ctx := context.TODO()
	run := runtime.NewRuntime()
	run = runtime.Compose(runtime.NewRuntime())

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
