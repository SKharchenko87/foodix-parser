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
	RequestDelay   int    `yaml:"request_delay"`
}

type Config struct {
	Log     Log            `yaml:"log"`
	Sources []SourceConfig `yaml:"sources"`
	Store   DB             `yaml:"db"`
}

type DB struct {
	Name      string `yaml:"name"`
	DSN       string `yaml:"dsn"`
	BatchSize int    `yaml:"batch_size"`
}

type Log struct {
	Level  string `yaml:"level,omitempty"`
	Format string `yaml:"format,omitempty"`
}

func LoadConfig(filePath string) (Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Config{}, fmt.Errorf("error reading config file %s %v", filePath, err)
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("error parsing config file %s %v", filePath, err)
	}

	return cfg, nil
}
