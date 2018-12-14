package cmd

import (
	"context"
	"github.com/gofunct/grpc12f/example/todo"
	"github.com/gofunct/grpc12f/runtime"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"log"
)

// grpcCmd represents the grpc command
var grpcCmd = &cobra.Command{
	Use:   "grpc",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()
		run, err := runtime.NewRuntime()
		if err != nil {
			log.Fatal("failed to create runtime", zap.Error(err))
		}
		defer run.Shutdown(ctx)
		run.Store.CreateTable(todo.Todo{}, nil)
		todo.RegisterTodoServiceServer(run.Server, &todo.Store{DB: run.Store})
		if err = run.Serve(ctx); err != nil {
			log.Fatal("failed to serve grpc", zap.Error(err))
		}
	},
}
