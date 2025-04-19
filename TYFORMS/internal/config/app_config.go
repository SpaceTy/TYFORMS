package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config holds all application configuration
type Config struct {
	Database struct {
		Path string `yaml:"path"`
	} `yaml:"database"`
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Admin struct {
		Password string `yaml:"password"`
	} `yaml:"admin"`
}

// DefaultConfig returns a new Config with default values
func DefaultConfig() *Config {
	cfg := &Config{}

	// Set default database path in the current directory
	cfg.Database.Path = "applications.db"

	// Default server port
	cfg.Server.Port = 5099

	// Default admin password (should be changed in production)
	cfg.Admin.Password = "admin123"

	return cfg
}

// LoadConfig loads the configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	// For now, we'll just return the default config
	// In a real application, you would load from a YAML file
	// using a library like github.com/spf13/viper

	cfg := DefaultConfig()

	// If config file exists, load it
	if _, err := os.Stat(configPath); err == nil {
		// TODO: Implement YAML loading
		fmt.Printf("Config file found at %s, but YAML loading not implemented yet\n", configPath)
	}

	return cfg, nil
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	// First try the current directory
	configPath := "config.yaml"
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	// Then try the user's home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath = filepath.Join(homeDir, ".tyforms", "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// Return default path if no config file found
	return "config.yaml"
}
