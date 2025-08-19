package storage

import (
	"github.com/SKharchenko87/foodix-parser/internal/models"
)

type DB interface {
	Close() error
	InsertProduct(product models.Product) error
	InsertProducts(products []models.Product) error
}
