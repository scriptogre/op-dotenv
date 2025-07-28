package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scriptogre/op-dotenv/internal/onepassword"
)

func TestGetFieldType(t *testing.T) {
	tests := []struct {
		fieldName string
		expected  string
	}{
		{"DATABASE_URL", "STRING"},
		{"API_KEY", "CONCEALED"},
		{"PASSWORD", "CONCEALED"},
		{"SMTP_PASS", "CONCEALED"},
		{"JWT_SECRET", "CONCEALED"},
		{"ACCESS_TOKEN", "CONCEALED"},
		{"OAUTH_AUTH", "CONCEALED"},
		{"DB_CREDENTIAL", "CONCEALED"},
		{"PASSWORD_HASH", "CONCEALED"},
		{"BCRYPT_SALT", "CONCEALED"},
		{"REDIS_HOST", "STRING"},
		{"PORT", "STRING"},
		{"DEBUG", "STRING"},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			result := getFieldType(tt.fieldName)
			if result != tt.expected {
				t.Errorf("getFieldType(%q) = %q, want %q", tt.fieldName, result, tt.expected)
			}
		})
	}
}

func TestParseEnvFileToItem(t *testing.T) {
	// Create a temporary .env file
	envContent := `# --------------------------------------------
# Test environment variables
# This is a test file
# --------------------------------------------

DATABASE_URL=postgres://localhost:5432/testdb
API_KEY=secret123

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379

# Email Settings
SMTP_HOST=smtp.gmail.com
SMTP_PASSWORD=email_secret
SMTP_USER=user@example.com
`

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envFile, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	// Parse the file
	item, err := ParseEnvFileToItem(envFile, "test-item")
	if err != nil {
		t.Fatalf("ParseEnvFileToItem failed: %v", err)
	}

	// Verify basic properties
	if item.Title != "test-item" {
		t.Errorf("Expected title 'test-item', got %q", item.Title)
	}

	// Check that we have the expected number of fields (7 env vars + notes)
	expectedFieldCount := 8
	if len(item.Fields) != expectedFieldCount {
		t.Errorf("Expected %d fields, got %d", expectedFieldCount, len(item.Fields))
	}

	// Verify notes field exists and contains expected content
	var notesField *onepassword.OnePasswordField
	for i := range item.Fields {
		if item.Fields[i].ID == "notesPlain" {
			notesField = &item.Fields[i]
			break
		}
	}

	if notesField == nil {
		t.Fatal("Notes field not found")
	}

	if !strings.Contains(notesField.Value, "Test environment variables") {
		t.Errorf("Notes should contain 'Test environment variables', got %q", notesField.Value)
	}

	// Verify field types
	fieldTypeTests := map[string]string{
		"DATABASE_URL":   "STRING",
		"API_KEY":        "CONCEALED",
		"REDIS_HOST":     "STRING",
		"REDIS_PORT":     "STRING",
		"SMTP_HOST":      "STRING",
		"SMTP_PASSWORD":  "CONCEALED",
		"SMTP_USER":      "STRING",
	}

	for _, field := range item.Fields {
		if field.ID == "notesPlain" {
			continue
		}

		expectedType, exists := fieldTypeTests[field.Label]
		if !exists {
			t.Errorf("Unexpected field: %s", field.Label)
			continue
		}

		if field.Type != expectedType {
			t.Errorf("Field %s: expected type %s, got %s", field.Label, expectedType, field.Type)
		}
	}

	// Verify sections
	sectionTests := map[string]string{
		"REDIS_HOST":     "Redis Configuration",
		"REDIS_PORT":     "Redis Configuration",
		"SMTP_HOST":      "Email Settings",
		"SMTP_PASSWORD":  "Email Settings",
		"SMTP_USER":      "Email Settings",
	}

	for _, field := range item.Fields {
		if expectedSection, exists := sectionTests[field.Label]; exists {
			if field.Section == nil {
				t.Errorf("Field %s should have section %q, but has no section", field.Label, expectedSection)
				continue
			}

			if sectionLabel, ok := field.Section["label"].(string); !ok || sectionLabel != expectedSection {
				t.Errorf("Field %s: expected section %q, got %q", field.Label, expectedSection, sectionLabel)
			}
		}
	}
}

func TestWriteItemToEnvFile(t *testing.T) {
	// Create a test item
	item := &onepassword.OnePasswordItem{
		Title: "test-item",
		Fields: []onepassword.OnePasswordField{
			{
				ID:    "notesPlain",
				Type:  "STRING",
				Label: "notesPlain",
				Value: "Test notes\nMultiple lines",
			},
			{
				Type:  "STRING",
				Label: "DATABASE_URL",
				Value: "postgres://localhost:5432/testdb",
			},
			{
				Type:  "CONCEALED",
				Label: "API_KEY",
				Value: "secret123",
			},
			{
				Type:  "STRING",
				Label: "REDIS_HOST",
				Value: "localhost",
				Section: map[string]interface{}{
					"label": "Redis Configuration",
				},
			},
			{
				Type:  "STRING",
				Label: "REDIS_PORT",
				Value: "6379",
				Section: map[string]interface{}{
					"label": "Redis Configuration",
				},
			},
		},
	}

	// Write to temporary file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	err := WriteItemToEnvFile(envFile, item)
	if err != nil {
		t.Fatalf("WriteItemToEnvFile failed: %v", err)
	}

	// Read the file back
	content, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("Failed to read generated .env file: %v", err)
	}

	contentStr := string(content)

	// Verify header with notes
	if !strings.Contains(contentStr, "Test notes") {
		t.Error("Generated .env should contain notes from item")
	}

	// Verify variables
	expectedVars := []string{
		"DATABASE_URL='postgres://localhost:5432/testdb'",
		"API_KEY='secret123'",
		"REDIS_HOST='localhost'",
		"REDIS_PORT='6379'",
	}

	for _, expectedVar := range expectedVars {
		if !strings.Contains(contentStr, expectedVar) {
			t.Errorf("Generated .env should contain: %s", expectedVar)
		}
	}

	// Verify section header
	if !strings.Contains(contentStr, "# Redis Configuration") {
		t.Error("Generated .env should contain section header '# Redis Configuration'")
	}
}