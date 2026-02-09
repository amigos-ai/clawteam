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

		url := fmt.Sprintf("http://localhost:%d/?token=%s", inst.Port, inst.GatewayToken)
		fmt.Printf("Instance '%s' created:\n  URL: %s\n", inst.Name, url)
		fmt.Println("\nOpen the URL in your browser, then run:")
		fmt.Printf("  clawteam pair %s\n", inst.Name)
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
		fmt.Fprintln(w, "NAME\tSTATUS\tURL")
		for _, inst := range instances {
			url := fmt.Sprintf("http://localhost:%d/?token=%s", inst.Port, inst.GatewayToken)
			if inst.Status != "running" {
				url = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", inst.Name, inst.Status, url)
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

var pairCmd = &cobra.Command{
	Use:   "pair <name>",
	Short: "Approve pending device pairings for an instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		specificID, _ := cmd.Flags().GetString("id")

		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		ctx := context.Background()

		pending, err := mgr.Pair(ctx, name)
		if err != nil {
			return err
		}

		if len(pending) == 0 {
			fmt.Println("No pending pairing requests")
			return nil
		}

		if specificID != "" {
			// Approve a specific device
			if err := mgr.ApprovePairing(ctx, name, specificID); err != nil {
				return err
			}
			fmt.Printf("Approved device %s\n", specificID)
			return nil
		}

		// Approve all pending devices
		for _, dev := range pending {
			fmt.Printf("Approving device %s", dev.ID)
			if dev.IP != "" {
				fmt.Printf(" (%s)", dev.IP)
			}
			fmt.Println()

			if err := mgr.ApprovePairing(ctx, name, dev.ID); err != nil {
				return fmt.Errorf("approve device %s: %w", dev.ID, err)
			}
		}

		fmt.Printf("\nApproved %d device(s). Refresh your browser to connect.\n", len(pending))
		return nil
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

	pairCmd.Flags().String("id", "", "Approve a specific device by request ID")
}
