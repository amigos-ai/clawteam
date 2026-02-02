package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the ClawTeam Docker image",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Building ClawTeam Docker image...")

		// First, ensure openclaw:local exists
		fmt.Println("Note: This requires 'openclaw:local' base image.")
		fmt.Println("Run OpenClaw's docker-setup.sh first if you haven't.")

		buildCmd := exec.Command("docker", "build", "-t", "clawteam:latest", ".")
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr

		if err := buildCmd.Run(); err != nil {
			return fmt.Errorf("docker build failed: %w", err)
		}

		fmt.Println("ClawTeam image built successfully!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
