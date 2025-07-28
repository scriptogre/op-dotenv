package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Projects map[string]ProjectConfig `json:"projects"`
}

type ProjectConfig struct {
	Vault string `json:"vault"`
	Item  string `json:"item"`
}

func LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return &Config{Projects: make(map[string]ProjectConfig)}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Projects: make(map[string]ProjectConfig)}, nil
		}
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	if config.Projects == nil {
		config.Projects = make(map[string]ProjectConfig)
	}

	return &config, nil
}

func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func (c *Config) GetVault(projectPath, defaultVault string) string {
	if project, exists := c.Projects[projectPath]; exists && project.Vault != "" {
		return project.Vault
	}
	return defaultVault
}

func (c *Config) GetItem(projectPath, defaultItem string) string {
	if project, exists := c.Projects[projectPath]; exists && project.Item != "" {
		return project.Item
	}
	return defaultItem
}

func (c *Config) SetVault(projectPath, vault string) {
	if c.Projects == nil {
		c.Projects = make(map[string]ProjectConfig)
	}
	
	project := c.Projects[projectPath]
	project.Vault = vault
	c.Projects[projectPath] = project
}

func (c *Config) SetItem(projectPath, item string) {
	if c.Projects == nil {
		c.Projects = make(map[string]ProjectConfig)
	}
	
	project := c.Projects[projectPath]
	project.Item = item
	c.Projects[projectPath] = project
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "op-dotenv", "config.json"), nil
}