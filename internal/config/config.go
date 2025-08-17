package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type SourceConfig struct {
	Name           string `yaml:"name"`
	URL            string `yaml:"url"`
	Timeout        int    `yaml:"timeout"`
	ProductPerPage int    `yaml:"product_per_page"`
}

type Config struct {
	Log     Log            `yaml:"log"`
	Sources []SourceConfig `yaml:"sources"`
	// ToDo
}

type Log struct {
	Level  string `yaml:"level,omitempty"`
	Format string `yaml:"format,omitempty"`
}

func LoadConfig(filePath string) (Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Config{}, fmt.Errorf("Error reading config file %s %v", filePath, err)
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("Error parsing config file %s %v", filePath, err)
	}

	return cfg, nil
}
