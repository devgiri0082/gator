package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// type ConfigMeta struct {
// 	file_path string
// }

type Config struct {
	DB_URL          string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func (c *Config) SetUser(name string) error {
	prev := c.CurrentUserName
	c.CurrentUserName = name
	if err := c.write(); err != nil {
		c.CurrentUserName = prev
		return err
	}
	return nil
}

func (c *Config) Getuser() string {
	username := c.CurrentUserName
	return username
}

func (c *Config) write() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("Error getting home: %w", err)
	}
	fullPath := filepath.Join(home, configFileName)
	enData, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("Error marshal data: %w", err)
	}
	err = os.WriteFile(fullPath, enData, 0666)
	if err != nil {
		return fmt.Errorf("Error writing: %w", err)
	}
	return nil
}

func Read() (*Config, error) {
	var newConfig Config
	home, err := os.UserHomeDir()
	if err != nil {
		return &newConfig, fmt.Errorf("uable to get homepath: %w", err)
	}
	fullPath := filepath.Join(home, configFileName)
	bytes, err := os.ReadFile(fullPath)
	if err != nil {
		return &newConfig, fmt.Errorf("error reading path: path: %s,   error: %w", fullPath, err)
	}
	json.Unmarshal(bytes, &newConfig)
	return &newConfig, nil

}
