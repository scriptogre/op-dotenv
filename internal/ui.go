package internal

import (
	"fmt"
	"os"

	"github.com/scriptogre/op-dotenv/internal/onepassword"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
)

// Color formatting functions
func Bold(text string) string {
	return colorBold + text + colorReset
}

func Green(text string) string {
	return colorGreen + text + colorReset
}

func Yellow(text string) string {
	return colorYellow + text + colorReset
}

func Red(text string) string {
	return colorRed + text + colorReset
}

// ConfirmOverwrite prompts user to confirm overwriting an existing item
func ConfirmOverwrite(itemType, name, location string) bool {
	fmt.Printf("\n%s %s '%s' exists in %s. Overwrite? (y/n): ",
		Yellow("⚠"), itemType, Bold(name), Bold(location))

	var response string
	fmt.Scanln(&response)

	if response != "y" && response != "Y" {
		fmt.Println("Operation cancelled.")
		return false
	}

	return true
}

// ShowSuccess displays a success message
func ShowSuccess(action, source, destination string) {
	fmt.Printf("\n💾 %s %s as %s.\n", action, Bold(source), Bold(destination))
}

// ShowError displays an error message to stderr
func ShowError(message string) {
	fmt.Fprintln(os.Stderr, message)
}

// ShowDependencyError displays styled dependency error messages
func ShowDependencyError(err error) {
	errorMsg := err.Error()

	if contains(errorMsg, "not authenticated") {
		fmt.Fprintf(os.Stderr, "%s %s\n", Red("🔐"), "1Password CLI not authenticated. Run "+Bold("op signin")+" first.")
	} else if contains(errorMsg, "not found") {
		fmt.Fprintf(os.Stderr, "%s %s\n", Red("🚫"), "1Password CLI not found")
		fmt.Fprintf(os.Stderr, "Install from: %s\n", Bold("https://developer.1password.com/docs/cli/get-started/"))
	} else {
		ShowError(errorMsg)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 1; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// HandleVaultNotFound provides interactive vault selection when a vault is not found
func HandleVaultNotFound(vaultName string) (string, error) {
	fmt.Printf("\n%s Vault '%s' not found.\n\n", Red("✗"), Bold(vaultName))

	// List available vaults
	vaults, err := onepassword.ListVaults()
	if err != nil {
		return "", fmt.Errorf("failed to list vaults: %w", err)
	}

	if len(vaults) > 0 {
		fmt.Printf("📁 %s\n", Bold("Available vaults:"))

		// Group vaults by name to handle duplicates
		vaultNames := make(map[string]int)
		for _, vault := range vaults {
			vaultNames[vault.Name]++
		}

		for _, vault := range vaults {
			if vaultNames[vault.Name] > 1 {
				// Show vault with ID for duplicates
				fmt.Printf("   %s %s (ID: %s)\n", Yellow("▸"), vault.Name, vault.ID)
			} else {
				// Show just the name for unique vaults
				fmt.Printf("   %s %s\n", Yellow("▸"), vault.Name)
			}
		}
		fmt.Println()
	}

	fmt.Printf("🛠️ %s\n", Bold("Choose an option:"))
	fmt.Printf("   %s %s\n", Green("1."), "Use an existing vault")
	fmt.Printf("   %s %s\n", Green("2."), "Create new vault")
	fmt.Printf("   %s %s\n", Red("3."), "Cancel")
	fmt.Printf("\n%s ", Bold("Enter choice (1/2/3):"))

	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		return selectExistingVault(vaults)
	case "2":
		return createNewVault()
	case "3":
		fmt.Println("\nOperation cancelled.")
		return "", nil
	default:
		return "", fmt.Errorf("invalid choice")
	}
}

// selectExistingVault handles selection of an existing vault
func selectExistingVault(vaults []onepassword.VaultInfo) (string, error) {
	if len(vaults) == 0 {
		fmt.Println("No vaults available.")
		return "", nil
	}

	fmt.Printf("\n📁 %s\n", Bold("Available vaults:"))
	for i, vault := range vaults {
		fmt.Printf("   %s %s\n", Yellow(fmt.Sprintf("%d.", i+1)), vault.Name)
	}
	fmt.Printf("\n%s ", Bold("Enter choice:"))

	var vaultChoice int
	fmt.Scanf("%d", &vaultChoice)

	if vaultChoice < 1 || vaultChoice > len(vaults) {
		return "", fmt.Errorf("invalid choice")
	}

	selectedVault := vaults[vaultChoice-1].Name
	fmt.Printf("\n✅ Using vault %s.\n", Bold(selectedVault))
	return selectedVault, nil
}

// createNewVault handles creation of a new vault
func createNewVault() (string, error) {
	fmt.Printf("\n📝 %s ", Bold("Enter vault name (leave empty for 'Environments'):"))
	var newVaultName string
	fmt.Scanln(&newVaultName)

	if newVaultName == "" {
		newVaultName = "Environments"
	}

	// Check if vault already exists
	err := ValidateVault(newVaultName)
	if err == nil {
		// Vault already exists
		fmt.Printf("\n✅ Vault %s already exists. Using existing vault.\n", Bold(newVaultName))
		return newVaultName, nil
	}

	// Vault doesn't exist, create it
	err = onepassword.CreateVault(newVaultName)
	if err != nil {
		return "", fmt.Errorf("failed to create vault: %w", err)
	}

	fmt.Printf("\n✨ Created vault %s.\n", Bold(newVaultName))
	return newVaultName, nil
}

// HandleItemNotFound provides interactive options when an item is not found
func HandleItemNotFound(vaultName, itemName string) (string, error) {
	fmt.Printf("\n%s Item '%s' not found in vault '%s'.\n\n", Red("✗"), Bold(itemName), Bold(vaultName))

	// List available items in the vault
	items, err := onepassword.ListItems(vaultName)
	if err != nil {
		return "", fmt.Errorf("failed to list items: %w", err)
	}

	if len(items) > 0 {
		fmt.Printf("📄 %s\n", Bold("Available items in this vault:"))
		for _, item := range items {
			fmt.Printf("   %s %s\n", Yellow("▸"), item.Title)
		}
		fmt.Println()
	}

	fmt.Printf("🛠️ %s\n", Bold("Choose an option:"))
	if len(items) > 0 {
		fmt.Printf("   %s %s\n", Green("1."), "Use an existing item")
		fmt.Printf("   %s %s\n", Red("2."), "Cancel")
		fmt.Printf("\n%s ", Bold("Enter choice (1/2):"))
	} else {
		fmt.Printf("   %s %s\n", Red("1."), "Cancel (no items found)")
		fmt.Printf("\n%s ", Bold("Enter choice (1):"))
	}

	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		if len(items) > 0 {
			return selectExistingItem(items)
		}
		fmt.Println("\nOperation cancelled.")
		return "", nil
	case "2":
		if len(items) > 0 {
			fmt.Println("\nOperation cancelled.")
			return "", nil
		}
		fallthrough
	default:
		return "", fmt.Errorf("invalid choice")
	}
}

// selectExistingItem handles selection of an existing item
func selectExistingItem(items []onepassword.ItemInfo) (string, error) {
	if len(items) == 0 {
		fmt.Println("No items available.")
		return "", nil
	}

	fmt.Printf("\n📄 %s\n", Bold("Available items:"))
	for i, item := range items {
		fmt.Printf("   %s %s\n", Yellow(fmt.Sprintf("%d.", i+1)), item.Title)
	}
	fmt.Printf("\n%s ", Bold("Enter choice:"))

	var itemChoice int
	fmt.Scanf("%d", &itemChoice)

	if itemChoice < 1 || itemChoice > len(items) {
		return "", fmt.Errorf("invalid choice")
	}

	selectedItem := items[itemChoice-1].Title
	fmt.Printf("\n✅ Using item %s.\n", Bold(selectedItem))
	return selectedItem, nil
}
