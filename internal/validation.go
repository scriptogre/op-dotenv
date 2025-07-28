package internal

import (
	"fmt"
	"os/exec"
)

// ValidateCliInstalled checks if 1Password CLI is installed
func ValidateCliInstalled() error {
	_, err := exec.LookPath("op")
	if err != nil {
		return fmt.Errorf("🚫 1Password CLI not found\nInstall from: https://developer.1password.com/docs/cli/get-started/")
	}
	return nil
}

// ValidateUserSignedIn checks if user is authenticated with 1Password CLI
func ValidateUserSignedIn() error {
	cmd := exec.Command("op", "whoami")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("🔐 1Password CLI not authenticated. Run 'op signin'")
	}
	return nil
}

// ValidateVault checks if a vault exists
func ValidateVault(vaultName string) error {
	cmd := exec.Command("op", "vault", "get", vaultName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("vault '%s' not found", vaultName)
	}
	return nil
}


