package main

import (
	"context"
	"github.com/gofunct/grpc12factor"
	"github.com/gofunct/grpc12factor/example/todo"
	"github.com/prometheus/common/log"
	"github.com/spf13/viper"
)

func main() {
	ctx := context.TODO()
	run := grpc12factor.NewRuntime()
	run = grpc12factor.Compose(grpc12factor.NewRuntime())
	defer run.Shutdown(ctx)

	err := todo.RegisterTodoServiceHandlerFromEndpoint(context.Background(), run.Gate, viper.GetString("grpc_port"), run.DialOpts)
	if err != nil {
		panic("Cannot serve http api")
	}
	run.Store.CreateTable(todo.Todo{}, nil)
	todo.RegisterTodoServiceServer(run.Server, &todo.Store{DB: run.Store})
	log.Fatal(run.Serve(ctx))
}
