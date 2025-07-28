package internal

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/scriptogre/op-dotenv/internal/onepassword"
)

// getFieldType determines if a field should be a password based on its name
func getFieldType(fieldName string) string {
	// Keywords that indicate a field should be hidden as a password
	passwordKeywords := []string{"PASSWORD", "PASS", "SECRET", "KEY", "TOKEN", "AUTH", "CREDENTIAL", "HASH", "SALT"}
	
	upperName := strings.ToUpper(fieldName)
	for _, keyword := range passwordKeywords {
		if strings.Contains(upperName, keyword) {
			return "CONCEALED"
		}
	}
	
	return "STRING"
}

// ParseEnvFileToItem reads a .env file and converts it to a OnePasswordItem structure
func ParseEnvFileToItem(filePath, itemTitle string) (*onepassword.OnePasswordItem, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	item := &onepassword.OnePasswordItem{
		Title:  itemTitle,
		Fields: []onepassword.OnePasswordField{},
	}

	scanner := bufio.NewScanner(file)
	currentSection := ""
	inHeader := false
	headerLines := []string{}

	// Regex patterns
	headerStartPattern := regexp.MustCompile(`^#\s*-+\s*$`)
	sectionPattern := regexp.MustCompile(`^#\s*(.+)\s*$`)
	varPattern := regexp.MustCompile(`^([A-Z_][A-Z0-9_]*)=(.*)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for header start (lines with dashes)
		if headerStartPattern.MatchString(line) {
			if !inHeader {
				inHeader = true
				continue
			} else {
				inHeader = false
				continue
			}
		}

		// If we're in header, collect notes
		if inHeader {
			if strings.HasPrefix(line, "#") {
				headerLines = append(headerLines, strings.TrimPrefix(strings.TrimSpace(line), "#"))
			}
			continue
		}

		// Check for section header
		if strings.HasPrefix(line, "#") && !headerStartPattern.MatchString(line) {
			matches := sectionPattern.FindStringSubmatch(line)
			if len(matches) > 1 {
				currentSection = strings.TrimSpace(matches[1])
				continue
			}
		}

		// Check for variable
		matches := varPattern.FindStringSubmatch(line)
		if len(matches) > 2 {
			key := matches[1]
			value := strings.Trim(matches[2], `'"`)

			field := onepassword.OnePasswordField{
				Type:  getFieldType(key),
				Label: key,
				Value: value,
			}

			// Add section if we're in one
			if currentSection != "" {
				field.Section = map[string]interface{}{
					"label": currentSection,
				}
			}

			item.Fields = append(item.Fields, field)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Add notes as a special field if present
	if len(headerLines) > 0 {
		notesField := onepassword.OnePasswordField{
			ID:    "notesPlain",
			Type:  "STRING",
			Label: "notesPlain",
			Value: strings.Join(headerLines, "\n"),
		}
		item.Fields = append(item.Fields, notesField)
	}

	return item, nil
}

// WriteItemToEnvFile converts a OnePasswordItem to a .env file
func WriteItemToEnvFile(filePath string, item *onepassword.OnePasswordItem) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Extract notes and organize fields by section
	var notes string
	sections := make(map[string][]onepassword.OnePasswordField)

	for _, field := range item.Fields {
		if field.ID == "notesPlain" {
			notes = field.Value
			continue
		}

		if field.Value == "" {
			continue
		}

		sectionName := ""
		if field.Section != nil {
			if label, ok := field.Section["label"].(string); ok {
				sectionName = label
			}
		}

		sections[sectionName] = append(sections[sectionName], field)
	}

	// Write header with notes if present
	if notes != "" {
		file.WriteString("# " + strings.Repeat("-", 44) + "\n")
		for _, line := range strings.Split(notes, "\n") {
			if strings.TrimSpace(line) != "" {
				file.WriteString("# " + line + "\n")
			}
		}
		file.WriteString("# " + strings.Repeat("-", 44) + "\n\n")
	}

	// Write ungrouped variables first (empty section key)
	if fields, exists := sections[""]; exists && len(fields) > 0 {
		for _, field := range fields {
			file.WriteString(fmt.Sprintf("%s='%s'\n", field.Label, field.Value))
		}
		file.WriteString("\n")
	}

	// Write sections
	for sectionName, fields := range sections {
		if sectionName == "" {
			continue // Already handled above
		}

		if len(fields) > 0 {
			file.WriteString(fmt.Sprintf("# %s\n", sectionName))
			for _, field := range fields {
				file.WriteString(fmt.Sprintf("%s='%s'\n", field.Label, field.Value))
			}
			file.WriteString("\n")
		}
	}

	return nil
}
