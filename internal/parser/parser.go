package parser

import (
	"errors"

	"github.com/SKharchenko87/foodix-parser/internal/config"
	"github.com/SKharchenko87/foodix-parser/internal/models"
)

type Parser interface {
	Parse() ([]models.Product, error)
	GetName() string
}

func NewParser(cfg config.SourceConfig) (Parser, error) {
	if cfg.Name == "calorizator" {
		return NewCalorizator(cfg)
	}
	return nil, errors.New("no parser found")
}
