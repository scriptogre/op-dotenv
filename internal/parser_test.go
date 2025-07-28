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

func TestCommentSpacingAndNewlines(t *testing.T) {
	tests := []struct {
		name        string
		envContent  string
		wantSpacing string
		wantLines   int
	}{
		{
			name: "single space in comments",
			envContent: `# --------------------------------------------
# This is a test comment
# --------------------------------------------

TEST_VAR=value`,
			wantSpacing: "# This is a test comment",
			wantLines:   1, // Should end with single newline
		},
		{
			name: "multiple spaces in comments should be normalized",
			envContent: `# --------------------------------------------
#   This has extra spaces
# --------------------------------------------

TEST_VAR=value`,
			wantSpacing: "# This has extra spaces",
			wantLines:   1,
		},
		{
			name: "flexible dash count",
			envContent: `# ---
# Short dashes
# ---

TEST_VAR=value`,
			wantSpacing: "# Short dashes",
			wantLines:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse .env to item
			tmpDir := t.TempDir()
			envFile := filepath.Join(tmpDir, ".env")
			err := os.WriteFile(envFile, []byte(tt.envContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test .env file: %v", err)
			}

			item, err := ParseEnvFileToItem(envFile, "test-item")
			if err != nil {
				t.Fatalf("ParseEnvFileToItem failed: %v", err)
			}

			// Write item back to .env
			outputFile := filepath.Join(tmpDir, "output.env")
			err = WriteItemToEnvFile(outputFile, item)
			if err != nil {
				t.Fatalf("WriteItemToEnvFile failed: %v", err)
			}

			// Read output content
			content, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			contentStr := string(content)

			// Check comment spacing
			if !strings.Contains(contentStr, tt.wantSpacing) {
				t.Errorf("Expected comment spacing %q, but content was:\n%s", tt.wantSpacing, contentStr)
			}

			// Check line count (should end with single newline, not multiple)
			lines := strings.Split(contentStr, "\n")
			emptyLinesAtEnd := 0
			for i := len(lines) - 1; i >= 0; i-- {
				if strings.TrimSpace(lines[i]) == "" {
					emptyLinesAtEnd++
				} else {
					break
				}
			}

			if emptyLinesAtEnd != tt.wantLines {
				t.Errorf("Expected %d empty lines at end, got %d. Content:\n%q", tt.wantLines, emptyLinesAtEnd, contentStr)
			}
		})
	}
}

func TestSectionOrderPreservation(t *testing.T) {
	tests := []struct {
		name       string
		envContent string
		wantOrder  []string
	}{
		{
			name: "original order: Email then Redis",
			envContent: `DATABASE_URL=postgres://localhost:5432/test

# Email Settings
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379`,
			wantOrder: []string{"Email Settings", "Redis Configuration"},
		},
		{
			name: "reversed order: Redis then Email",
			envContent: `DATABASE_URL=postgres://localhost:5432/test

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379

# Email Settings
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587`,
			wantOrder: []string{"Redis Configuration", "Email Settings"},
		},
		{
			name: "three sections in specific order",
			envContent: `# Database
DB_HOST=localhost

# Cache
CACHE_URL=redis://localhost

# API
API_KEY=secret`,
			wantOrder: []string{"Database", "Cache", "API"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse .env to item
			tmpDir := t.TempDir()
			envFile := filepath.Join(tmpDir, ".env")
			err := os.WriteFile(envFile, []byte(tt.envContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test .env file: %v", err)
			}

			item, err := ParseEnvFileToItem(envFile, "test-item")
			if err != nil {
				t.Fatalf("ParseEnvFileToItem failed: %v", err)
			}

			// Write item back to .env
			outputFile := filepath.Join(tmpDir, "output.env")
			err = WriteItemToEnvFile(outputFile, item)
			if err != nil {
				t.Fatalf("WriteItemToEnvFile failed: %v", err)
			}

			// Read output content and verify section order
			content, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			contentStr := string(content)
			lines := strings.Split(contentStr, "\n")

			// Find section headers in order
			var foundSections []string
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "#") && !strings.Contains(line, "---") {
					sectionName := strings.TrimSpace(strings.TrimPrefix(line, "#"))
					if sectionName != "" {
						foundSections = append(foundSections, sectionName)
					}
				}
			}

			// Verify order matches expected
			if len(foundSections) != len(tt.wantOrder) {
				t.Errorf("Expected %d sections, found %d: %v", len(tt.wantOrder), len(foundSections), foundSections)
				return
			}

			for i, expected := range tt.wantOrder {
				if i >= len(foundSections) || foundSections[i] != expected {
					t.Errorf("Section order mismatch at position %d: expected %q, got %q", i, expected, foundSections[i])
				}
			}
		})
	}
}

func TestRoundTripConsistency(t *testing.T) {
	originalEnv := `# --------------------------------------------
# My Application Environment
# --------------------------------------------

DATABASE_URL=postgres://localhost:5432/myapp
API_KEY=secret123

# Email Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_PASSWORD=email_secret

# Redis Settings
REDIS_HOST=localhost
REDIS_PORT=6379`

	tmpDir := t.TempDir()
	
	// Step 1: Write original .env file
	originalFile := filepath.Join(tmpDir, "original.env")
	err := os.WriteFile(originalFile, []byte(originalEnv), 0644)
	if err != nil {
		t.Fatalf("Failed to create original .env file: %v", err)
	}
	
	// Step 2: Parse to 1Password item
	item, err := ParseEnvFileToItem(originalFile, "test-item")
	if err != nil {
		t.Fatalf("Failed to parse original .env: %v", err)
	}
	
	// Step 3: Write back to .env file
	roundTripFile := filepath.Join(tmpDir, "roundtrip.env")
	err = WriteItemToEnvFile(roundTripFile, item)
	if err != nil {
		t.Fatalf("Failed to write round-trip .env: %v", err)
	}
	
	// Step 4: Parse round-trip file again
	item2, err := ParseEnvFileToItem(roundTripFile, "test-item")
	if err != nil {
		t.Fatalf("Failed to parse round-trip .env: %v", err)
	}
	
	// Step 5: Compare field types and values
	fieldMap1 := make(map[string]onepassword.OnePasswordField)
	fieldMap2 := make(map[string]onepassword.OnePasswordField)
	
	for _, field := range item.Fields {
		if field.ID != "notesPlain" {
			fieldMap1[field.Label] = field
		}
	}
	
	for _, field := range item2.Fields {
		if field.ID != "notesPlain" {
			fieldMap2[field.Label] = field
		}
	}
	
	// Verify same fields exist
	if len(fieldMap1) != len(fieldMap2) {
		t.Errorf("Field count mismatch: original %d, round-trip %d", len(fieldMap1), len(fieldMap2))
	}
	
	// Verify field types and values are preserved
	for label, field1 := range fieldMap1 {
		field2, exists := fieldMap2[label]
		if !exists {
			t.Errorf("Field %s missing in round-trip", label)
			continue
		}
		
		if field1.Type != field2.Type {
			t.Errorf("Field %s type mismatch: original %s, round-trip %s", label, field1.Type, field2.Type)
		}
		
		if field1.Value != field2.Value {
			t.Errorf("Field %s value mismatch: original %s, round-trip %s", label, field1.Value, field2.Value)
		}
		
		// Check section consistency
		section1 := ""
		section2 := ""
		if field1.Section != nil {
			if s, ok := field1.Section["label"].(string); ok {
				section1 = s
			}
		}
		if field2.Section != nil {
			if s, ok := field2.Section["label"].(string); ok {
				section2 = s
			}
		}
		
		if section1 != section2 {
			t.Errorf("Field %s section mismatch: original %s, round-trip %s", label, section1, section2)
		}
	}
}

func TestFieldTypeConsistency(t *testing.T) {
	envContent := `API_KEY=secret123
PASSWORD=mypass
JWT_TOKEN=jwt123
DATABASE_URL=postgres://localhost
REDIS_HOST=localhost
DEBUG=true`

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envFile, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	item, err := ParseEnvFileToItem(envFile, "test-item")
	if err != nil {
		t.Fatalf("ParseEnvFileToItem failed: %v", err)
	}

	expectedTypes := map[string]string{
		"API_KEY":      "CONCEALED",
		"PASSWORD":     "CONCEALED", 
		"JWT_TOKEN":    "CONCEALED",
		"DATABASE_URL": "STRING",
		"REDIS_HOST":   "STRING",
		"DEBUG":        "STRING",
	}

	for _, field := range item.Fields {
		if field.ID == "notesPlain" {
			continue
		}
		
		expectedType, exists := expectedTypes[field.Label]
		if !exists {
			t.Errorf("Unexpected field: %s", field.Label)
			continue
		}
		
		if field.Type != expectedType {
			t.Errorf("Field %s: expected type %s, got %s", field.Label, expectedType, field.Type)
		}
	}
}

func TestSectionReorderScenario(t *testing.T) {
	// This test simulates the real-world scenario where user changes section order in .env
	// and expects the change to be reflected in 1Password
	
	originalOrder := `DATABASE_URL=postgres://localhost:5432/test

# Email Settings  
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379`

	reorderedEnv := `DATABASE_URL=postgres://localhost:5432/test

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379

# Email Settings
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587`

	tmpDir := t.TempDir()
	
	// Step 1: Parse original order
	originalFile := filepath.Join(tmpDir, "original.env")
	err := os.WriteFile(originalFile, []byte(originalOrder), 0644)
	if err != nil {
		t.Fatalf("Failed to create original .env file: %v", err)
	}

	originalItem, err := ParseEnvFileToItem(originalFile, "test-item")
	if err != nil {
		t.Fatalf("Failed to parse original .env: %v", err)
	}

	// Step 2: Parse reordered version
	reorderedFile := filepath.Join(tmpDir, "reordered.env")
	err = os.WriteFile(reorderedFile, []byte(reorderedEnv), 0644)
	if err != nil {
		t.Fatalf("Failed to create reordered .env file: %v", err)
	}

	reorderedItem, err := ParseEnvFileToItem(reorderedFile, "test-item")
	if err != nil {
		t.Fatalf("Failed to parse reordered .env: %v", err)
	}

	// Step 3: Verify both have same fields but different section order
	originalOutput := filepath.Join(tmpDir, "original_output.env")
	err = WriteItemToEnvFile(originalOutput, originalItem)
	if err != nil {
		t.Fatalf("Failed to write original output: %v", err)
	}

	reorderedOutput := filepath.Join(tmpDir, "reordered_output.env")
	err = WriteItemToEnvFile(reorderedOutput, reorderedItem)
	if err != nil {
		t.Fatalf("Failed to write reordered output: %v", err)
	}

	// Read both outputs
	originalContent, err := os.ReadFile(originalOutput)
	if err != nil {
		t.Fatalf("Failed to read original output: %v", err)
	}

	reorderedContent, err := os.ReadFile(reorderedOutput)
	if err != nil {
		t.Fatalf("Failed to read reordered output: %v", err)
	}

	// Extract section order from both
	originalSections := extractSectionOrder(string(originalContent))
	reorderedSections := extractSectionOrder(string(reorderedContent))

	// Verify they have different orders
	expectedOriginal := []string{"Email Settings", "Redis Configuration"}
	expectedReordered := []string{"Redis Configuration", "Email Settings"}

	if !slicesEqual(originalSections, expectedOriginal) {
		t.Errorf("Original section order incorrect: got %v, want %v", originalSections, expectedOriginal)
	}

	if !slicesEqual(reorderedSections, expectedReordered) {
		t.Errorf("Reordered section order incorrect: got %v, want %v", reorderedSections, expectedReordered)
	}

	// Verify section order is different between the two
	if slicesEqual(originalSections, reorderedSections) {
		t.Error("Section order should be different between original and reordered, but they are the same")
	}
}

// Helper function to extract section order from .env content
func extractSectionOrder(content string) []string {
	var sections []string
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") && !strings.Contains(line, "---") {
			sectionName := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			if sectionName != "" {
				sections = append(sections, sectionName)
			}
		}
	}
	
	return sections
}

// Helper function to compare string slices
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}