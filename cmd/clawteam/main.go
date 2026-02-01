package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "clawteam",
	Short: "OpenClaw orchestration tool",
	Long:  "Manage isolated OpenClaw instances in Docker containers",
}

func init() {
	rootCmd.AddCommand(vaultCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
