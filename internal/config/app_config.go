package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
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
	cfg := DefaultConfig()

	// If config file exists, load it
	if _, err := os.Stat(configPath); err == nil {
		// Read the config file
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}

		// Parse YAML into config struct
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("error parsing config file: %w", err)
		}
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
