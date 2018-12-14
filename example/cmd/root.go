package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "grpc12f",
	Short: "grpc12f is a highly configurable grpc server- config.yaml holds all of the config options available for your server",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(grpcCmd)

	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}