package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/amigos-ai/clawteam/internal/instance"
	"github.com/spf13/cobra"
)

func getManager() (*instance.Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home directory: %w", err)
	}
	return instance.NewManager(filepath.Join(home, ".clawteam"))
}

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new OpenClaw instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		anthropic, _ := cmd.Flags().GetString("anthropic")
		openai, _ := cmd.Flags().GetString("openai")
		ssh, _ := cmd.Flags().GetString("ssh")
		gitName, _ := cmd.Flags().GetString("git-name")
		gitEmail, _ := cmd.Flags().GetString("git-email")
		persistence, _ := cmd.Flags().GetString("persistence")
		port, _ := cmd.Flags().GetInt("port")

		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		inst, err := mgr.Create(context.Background(), instance.CreateOptions{
			Name:         name,
			AnthropicRef: anthropic,
			OpenAIRef:    openai,
			SSHRef:       ssh,
			GitName:      gitName,
			GitEmail:     gitEmail,
			Persistence:  instance.PersistenceLevel(persistence),
			Port:         port,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Instance '%s' running at http://localhost:%d\n", inst.Name, inst.Port)
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all instances",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		instances, err := mgr.List()
		if err != nil {
			return err
		}

		if len(instances) == 0 {
			fmt.Println("No instances")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSTATUS\tPORT\tPERSISTENCE")
		for _, inst := range instances {
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", inst.Name, inst.Status, inst.Port, inst.Persistence)
		}
		w.Flush()

		return nil
	},
}

var startCmd = &cobra.Command{
	Use:   "start <name>",
	Short: "Start a stopped instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		if err := mgr.Start(context.Background(), args[0]); err != nil {
			return err
		}

		fmt.Printf("Instance '%s' started\n", args[0])
		return nil
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop a running instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		if err := mgr.Stop(context.Background(), args[0]); err != nil {
			return err
		}

		fmt.Printf("Instance '%s' stopped\n", args[0])
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete an instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		if err := mgr.Delete(context.Background(), args[0]); err != nil {
			return err
		}

		fmt.Printf("Instance '%s' deleted\n", args[0])
		return nil
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs <name>",
	Short: "View instance logs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		follow, _ := cmd.Flags().GetBool("follow")

		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		return mgr.Logs(context.Background(), args[0], follow)
	},
}

func init() {
	createCmd.Flags().String("anthropic", "", "Vault ref for Anthropic API key")
	createCmd.Flags().String("openai", "", "Vault ref for OpenAI API key")
	createCmd.Flags().String("ssh", "", "Vault ref for SSH key")
	createCmd.Flags().String("git-name", "", "Git user.name")
	createCmd.Flags().String("git-email", "", "Git user.email")
	createCmd.Flags().String("persistence", "full", "Persistence level (full, memory, minimal)")
	createCmd.Flags().Int("port", 0, "Specific port (default: auto-assign)")

	logsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
}
