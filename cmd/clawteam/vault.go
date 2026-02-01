package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/amigos-ai/clawteam/internal/vault"
	"github.com/spf13/cobra"
)

func getVaultStorage() (*vault.Storage, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home directory: %w", err)
	}
	return vault.NewStorage(filepath.Join(home, ".clawteam", "vault")), nil
}

var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Manage credentials",
}

var vaultAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a credential to the vault",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		credType, _ := cmd.Flags().GetString("type")
		value, _ := cmd.Flags().GetString("value")
		file, _ := cmd.Flags().GetString("file")

		// Validate credential type
		if credType != string(vault.TypeAPIKey) && credType != string(vault.TypeSSHKey) {
			return fmt.Errorf("invalid credential type %q: must be %q or %q", credType, vault.TypeAPIKey, vault.TypeSSHKey)
		}

		if value == "" && file == "" {
			return fmt.Errorf("must provide --value or --file")
		}

		if file != "" {
			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}
			value = string(data)
		}

		cred := &vault.Credential{
			Name:      name,
			Type:      vault.CredentialType(credType),
			Value:     value,
			CreatedAt: time.Now(),
		}

		store, err := getVaultStorage()
		if err != nil {
			return err
		}
		if err := store.Save(cred); err != nil {
			return err
		}

		fmt.Printf("Added credential '%s' (%s)\n", name, credType)
		return nil
	},
}

var vaultListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getVaultStorage()
		if err != nil {
			return err
		}
		creds, err := store.List()
		if err != nil {
			return err
		}

		if len(creds) == 0 {
			fmt.Println("No credentials in vault")
			return nil
		}

		for _, c := range creds {
			fmt.Printf("%-20s %s\n", c.Name, c.Type)
		}
		return nil
	},
}

var vaultRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a credential from the vault",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		store, err := getVaultStorage()
		if err != nil {
			return err
		}
		if err := store.Remove(name); err != nil {
			return err
		}
		fmt.Printf("Removed credential '%s'\n", name)
		return nil
	},
}

func init() {
	vaultAddCmd.Flags().StringP("type", "t", "api-key", "Credential type (api-key, ssh-key)")
	vaultAddCmd.Flags().StringP("value", "v", "", "Credential value")
	vaultAddCmd.Flags().StringP("file", "f", "", "Read value from file")

	vaultCmd.AddCommand(vaultAddCmd)
	vaultCmd.AddCommand(vaultListCmd)
	vaultCmd.AddCommand(vaultRemoveCmd)
}
