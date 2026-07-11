package cmd

import (
	"github.com/spf13/cobra"
)

var serverAddr string

var rootCmd = &cobra.Command{
	Use:   "myfitnesspal",
	Short: "Track foods and daily macros",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&serverAddr, "addr", "localhost:50051", "gRPC server address")
	rootCmd.AddCommand(foodCmd)
	rootCmd.AddCommand(logCmd)
}
