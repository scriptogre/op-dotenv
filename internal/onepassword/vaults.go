package onepassword

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// ListVaults returns all available vaults
func ListVaults() ([]VaultInfo, error) {
	cmd := exec.Command("op", "vault", "list", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var vaults []VaultInfo
	err = json.Unmarshal(output, &vaults)
	if err != nil {
		return nil, err
	}

	return vaults, nil
}

// CreateVault creates a new vault
func CreateVault(vaultName string) error {
	cmd := exec.Command("op", "vault", "create", vaultName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create vault: %s", string(output))
	}
	return nil
}

// GetVaultIdentifier returns the vault ID if there are multiple vaults with the same name,
// otherwise returns the vault name
func GetVaultIdentifier(vaultName string) (string, error) {
	vaults, err := ListVaults()
	if err != nil {
		return "", fmt.Errorf("failed to list vaults: %w", err)
	}
	
	var matchingVaults []VaultInfo
	for _, v := range vaults {
		if v.Name == vaultName {
			matchingVaults = append(matchingVaults, v)
		}
	}
	
	if len(matchingVaults) == 0 {
		return "", fmt.Errorf("vault '%s' not found", vaultName)
	}
	
	if len(matchingVaults) == 1 {
		// Only one vault with this name, can use name
		return vaultName, nil
	}
	
	// Multiple vaults with same name, need to pick one or ask user
	// For now, return the first one's ID
	return matchingVaults[0].ID, nil
}