package onepassword

// OnePasswordItem represents a 1Password item structure
type OnePasswordItem struct {
	ID     string                 `json:"id"`
	Title  string                 `json:"title"`
	Fields []OnePasswordField     `json:"fields"`
	Vault  map[string]interface{} `json:"vault"`
}

// OnePasswordField represents a field within a 1Password item
type OnePasswordField struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Label   string                 `json:"label"`
	Value   string                 `json:"value"`
	Section map[string]interface{} `json:"section,omitempty"`
}

// VaultInfo represents vault information from 1Password
type VaultInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ItemInfo represents basic item information from 1Password
type ItemInfo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}