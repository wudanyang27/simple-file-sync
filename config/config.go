package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the configuration structure for simple-file-sync
type Config struct {
	// Client configuration
	Mode        string `toml:"mode"`
	LocalDir    string `toml:"local_dir"`
	RemoteDir   string `toml:"remote_dir"`
	ServerAddr  string `toml:"server_addr"`
	ServerToken string `toml:"server_token"`

	// Additional configuration (from the deleted config file)
	Ignore       []string          `toml:"ignore"`
	PathMappings []string          `toml:"path_mappings"`
	
	// Internal field to track config file path
	ConfigFilePath string `toml:"-"`
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	return &Config{
		Mode:        "all",
		ServerToken: "kfcvme50",
	}
}

// LoadConfig loads configuration from a TOML file
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", configPath)
	}

	// Parse TOML file
	if _, err := toml.DecodeFile(configPath, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Store the config file path
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for config file: %w", err)
	}
	config.ConfigFilePath = absPath

	// Default local_dir to config file directory if not specified
	if config.LocalDir == "" {
		config.LocalDir = filepath.Dir(absPath)
	}

	return config, nil
}

// FindConfigFile looks for a config file in common locations
func FindConfigFile() (string, error) {
	// Look for config files in order of preference
	candidates := []string{
		"simple-file-sync.toml",
		".simple-file-sync.toml",
		"config/simple-file-sync.toml",
	}

	// Also check home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(homeDir, ".simple-file-sync.toml"))
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("no config file found in common locations")
}