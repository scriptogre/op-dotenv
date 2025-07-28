package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/scriptogre/op-dotenv/internal/onepassword"
)

// App represents the application with its dependencies
type App struct {
	config *Config
}

// NewApp creates a new application instance
func NewApp() (*App, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &App{
		config: config,
	}, nil
}

// Push uploads a .env file to 1Password
func (a *App) Push(filePath, vault, item string, force bool) error {
	// Validate dependencies first
	if err := ValidateCliInstalled(); err != nil {
		ShowDependencyError(err)
		os.Exit(1)
	}

	if err := ValidateUserSignedIn(); err != nil {
		ShowDependencyError(err)
		os.Exit(1)
	}

	// Determine target vault and item
	targetVault, targetItem, err := a.resolveTarget(vault, item)
	if err != nil {
		return err
	}

	// Try to resolve vault to ID (handles existence check)
	vaultID, err := onepassword.GetVaultIdentifier(targetVault)
	if err != nil {
		// Vault not found - let user choose
		selectedVault, err := HandleVaultNotFound(targetVault)
		if err != nil {
			return err
		}
		if selectedVault == "" {
			return nil // User cancelled
		}
		// Update targetVault to use selected vault
		targetVault = selectedVault
		// Get ID for selected vault
		vaultID, err = onepassword.GetVaultIdentifier(selectedVault)
		if err != nil {
			return fmt.Errorf("failed to resolve selected vault: %w", err)
		}
	}

	// Parse .env file to 1Password item
	parsedItem, err := ParseEnvFileToItem(filePath, targetItem)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	// Check if item exists and confirm overwrite
	if onepassword.ItemExists(vaultID, targetItem) {
		if !force && !ConfirmOverwrite("Item", targetItem, "vault '"+targetVault+"'") {
			return nil
		}
	}

	// Extract notes and fields from item
	notes := ""
	var fields []onepassword.OnePasswordField

	for _, field := range parsedItem.Fields {
		if field.ID == "notesPlain" {
			notes = field.Value
		} else {
			fields = append(fields, field)
		}
	}

	// Create or update the item
	if onepassword.ItemExists(vaultID, targetItem) {
		existingItem, err := onepassword.GetItemByName(vaultID, targetItem)
		if err != nil {
			return err
		}
		err = onepassword.UpdateItemFields(existingItem.ID, notes, fields)
	} else {
		err = onepassword.CreateItemFromFields(vaultID, targetItem, notes, fields)
	}

	if err != nil {
		return fmt.Errorf("failed to update 1Password item: %w", err)
	}

	// Save the vault and item choices for future use
	workingDir, _ := os.Getwd()
	a.config.SetVault(workingDir, targetVault)
	a.config.SetItem(workingDir, targetItem)
	a.config.Save() // Ignore error - not critical

	ShowSuccess("Saved", filePath, targetVault+"/"+targetItem+" in 1Password")
	return nil
}

// Pull downloads a 1Password item to a .env file
func (a *App) Pull(filePath, vault, item string) error {
	// Validate dependencies first
	if err := ValidateCliInstalled(); err != nil {
		ShowDependencyError(err)
		os.Exit(1)
	}

	if err := ValidateUserSignedIn(); err != nil {
		ShowDependencyError(err)
		os.Exit(1)
	}

	// Determine target vault and item
	targetVault, targetItem, err := a.resolveTarget(vault, item)
	if err != nil {
		return err
	}

	// Try to resolve vault to ID (handles existence check)
	vaultID, err := onepassword.GetVaultIdentifier(targetVault)
	if err != nil {
		// Vault not found - let user choose
		selectedVault, err := HandleVaultNotFound(targetVault)
		if err != nil {
			return err
		}
		if selectedVault == "" {
			return nil // User cancelled
		}
		// Update targetVault to use selected vault
		targetVault = selectedVault
		// Get ID for selected vault
		vaultID, err = onepassword.GetVaultIdentifier(selectedVault)
		if err != nil {
			return fmt.Errorf("failed to resolve selected vault: %w", err)
		}
	}

	// Get item from 1Password
	opItem, err := onepassword.GetItemByName(vaultID, targetItem)
	if err != nil {
		// Item not found - let user choose
		selectedItem, err := HandleItemNotFound(targetVault, targetItem)
		if err != nil {
			return err
		}
		if selectedItem == "" {
			return nil // User cancelled
		}
		// Update targetItem to use selected item
		targetItem = selectedItem
		// Get the selected item
		opItem, err = onepassword.GetItemByName(vaultID, selectedItem)
		if err != nil {
			return fmt.Errorf("failed to get selected item: %w", err)
		}
	}

	// Check if file exists and confirm overwrite
	if _, err := os.Stat(filePath); err == nil {
		if !ConfirmOverwrite("File", filePath, "local filesystem") {
			return nil
		}
	}

	// Write item to .env file
	err = WriteItemToEnvFile(filePath, opItem)
	if err != nil {
		return fmt.Errorf("failed to generate %s: %w", filePath, err)
	}

	// Save the vault and item choices for future use
	workingDir, _ := os.Getwd()
	a.config.SetVault(workingDir, targetVault)
	a.config.SetItem(workingDir, targetItem)
	a.config.Save() // Ignore error - not critical

	ShowSuccess("Saved", targetVault+"/"+targetItem, filePath+" from 1Password")
	return nil
}

// resolveTarget determines the target vault and item names
func (a *App) resolveTarget(vault, item string) (string, string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	targetVault := vault
	targetItem := item

	if targetVault == "" {
		targetVault = a.config.GetVault(workingDir, "Environments")
	}
	if targetItem == "" {
		targetItem = a.config.GetItem(workingDir, filepath.Base(workingDir))
	}

	return targetVault, targetItem, nil
}

// Clean removes all configuration data
func (a *App) Clean() error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("No configuration data found.")
		return nil
	}

	// Remove the config file
	if err := os.Remove(configPath); err != nil {
		return fmt.Errorf("failed to remove config file: %w", err)
	}

	// Try to remove the config directory if it's empty
	configDir := filepath.Dir(configPath)
	os.Remove(configDir) // Ignore error - directory might not be empty

	fmt.Println("Configuration data removed successfully.")
	return nil
}
