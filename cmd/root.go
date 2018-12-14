package cmd

import "github.com/spf13/cobra"

func New() *cobra.Command {
	return &cobra.Command{
		Use:   "runtime",
		Short: "runtime is an golang engine to help power your grpc microservices",
	}
}
