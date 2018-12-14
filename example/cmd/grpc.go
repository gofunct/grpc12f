package cmd

import (
	"context"
	"github.com/gofunct/grpc12factor"
	"github.com/gofunct/grpc12factor/example/todo"
	"github.com/spf13/cobra"
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
		run := grpc12factor.NewRuntime()
		run = grpc12factor.Compose(grpc12factor.NewRuntime())
		defer run.Shutdown(ctx)

		run.Store.CreateTable(todo.Todo{}, nil)
		todo.RegisterTodoServiceServer(run.Server, &todo.Store{DB: run.Store})
		log.Fatal(run.Serve(ctx))
	},
}

func init() {
	rootCmd.AddCommand(grpcCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// grpcCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// grpcCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
