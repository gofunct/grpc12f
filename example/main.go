package main

import (
	"context"
	"fmt"
	"github.com/gofunct/runtime"
)

func main() {
	run := runtime.Compose(
		runtime.WithLogger(),
		runtime.WithTracer(),
		runtime.WithRouter(),
		runtime.WithStore(),
		runtime.WithRootCmd(),
		runtime.WithServer(true),
	)

	api := newDemoServer()
	RegisterDemoServiceServer(run.Server, api)
	run.Serve()
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
