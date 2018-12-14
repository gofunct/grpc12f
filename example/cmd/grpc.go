package cmd

import (
	"context"
	"github.com/gofunct/grpc12factor"
	"github.com/gofunct/grpc12factor/example/todo"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"log"
)

// grpcCmd represents the grpc command
var grpcCmd = &cobra.Command{
	Use:   "grpc",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()
		run, err := grpc12factor.NewRuntime()
		if err != nil {
			log.Fatal("failed to create runtime", zap.Error(err))
		}
		defer run.Shutdown(ctx)
		run.Store.CreateTable(todo.Todo{}, nil)
		todo.RegisterTodoServiceServer(run.Server, &todo.Store{DB: run.Store})
		if err = run.Serve(ctx); err != nil {
			run.Log.Fatal("failed to serve grpc", zap.Error(err))
		}
	},
}

func init() {
	rootCmd.AddCommand(grpcCmd)
}
