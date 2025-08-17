package parser

import (
	"github.com/SKharchenko87/foodix-parser/internal/models"
)

type Parser interface {
	Parse() ([]models.Product, error)
}
