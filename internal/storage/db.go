package storage

import (
	"errors"

	"github.com/SKharchenko87/foodix-parser/internal/config"
	"github.com/SKharchenko87/foodix-parser/internal/models"
)

type DB interface {
	Close() error
	InsertProduct(product models.Product) error
	InsertProducts(products []models.Product) error
}

func NewStore(cfg config.DB) (DB, error) {
	if cfg.Name == "postgres" {
		return NewPostgres(cfg)
	}
	return nil, errors.New("unsupported database")
}
