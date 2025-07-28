package onepassword

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// GetItemByName retrieves a 1Password item by name from a vault
func GetItemByName(vault, itemName string) (*OnePasswordItem, error) {
	cmd := exec.Command("op", "item", "get", itemName, "--vault", vault, "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("item '%s' not found in vault '%s'", itemName, vault)
	}

	var item OnePasswordItem
	err = json.Unmarshal(output, &item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// ItemExists checks if an item exists in the specified vault
func ItemExists(vault, itemName string) bool {
	_, err := GetItemByName(vault, itemName)
	return err == nil
}

// CreateItemFromFields creates a new 1Password item with the given fields
func CreateItemFromFields(vault, itemName, notes string, fields []OnePasswordField) error {
	args := []string{"item", "create", "--category", "Secure Note", "--title", itemName, "--vault", vault}

	// Add notes if present
	if notes != "" {
		args = append(args, fmt.Sprintf("notesPlain=%s", notes))
	}

	// Add fields using assignment syntax (which works properly for sections and field types)
	for _, field := range fields {
		if field.ID == "notesPlain" {
			continue // Already handled above
		}

		var fieldAssignment string
		if field.Section != nil {
			if sectionLabel, ok := field.Section["label"].(string); ok {
				// Field with section: section.field[type]=value  
				fieldAssignment = fmt.Sprintf("%s.%s[%s]=%s", sectionLabel, field.Label, field.Type, field.Value)
			}
		} else {
			// Field without section: field[type]=value
			fieldAssignment = fmt.Sprintf("%s[%s]=%s", field.Label, field.Type, field.Value)
		}
		
		args = append(args, fieldAssignment)
	}

	cmd := exec.Command("op", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create item: %s", string(output))
	}

	return nil
}

// UpdateItemFields updates an existing 1Password item with new fields
func UpdateItemFields(itemID, notes string, fields []OnePasswordField) error {
	args := []string{"item", "edit", itemID}

	// Update notes if present
	if notes != "" {
		args = append(args, fmt.Sprintf("notesPlain=%s", notes))
	}

	// Update fields
	for _, field := range fields {
		if field.Section != nil {
			if sectionLabel, ok := field.Section["label"].(string); ok {
				// Field with section: section.field=value
				args = append(args, fmt.Sprintf("%s.%s=%s", sectionLabel, field.Label, field.Value))
			}
		} else {
			// Field without section
			args = append(args, fmt.Sprintf("%s=%s", field.Label, field.Value))
		}
	}

	cmd := exec.Command("op", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update item: %s", string(output))
	}

	return nil
}

// ListItems returns all items in a vault
func ListItems(vault string) ([]ItemInfo, error) {
	cmd := exec.Command("op", "item", "list", "--vault", vault, "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list items in vault '%s': %w", vault, err)
	}

	var items []ItemInfo
	err = json.Unmarshal(output, &items)
	if err != nil {
		return nil, err
	}

	return items, nil
}