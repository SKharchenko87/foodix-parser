package storage

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/SKharchenko87/foodix-parser/internal/config"
	"github.com/SKharchenko87/foodix-parser/internal/models"
	_ "github.com/lib/pq" // postgres driver
)

type Postgres struct {
	db  *sql.DB
	cfg config.DB
}

func NewPostgres(cfg config.DB) (DB, error) {
	os.Unsetenv("PGLOCALEDIR")
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, err
	}
	res := new(Postgres)
	res.db = db
	res.cfg = cfg
	return res, nil
}

func (p *Postgres) Close() error {
	return p.db.Close()
}

func (p *Postgres) InsertProduct(product models.Product) (err error) {
	_, err = p.db.Exec(
		"INSERT INTO product(name, protein, fat, carbohydrate, kcal) VALUES($1, $2, $3, $4, $5)",
		product.Name, product.Protein, product.Fat, product.Carbohydrate, product.Kcal,
	)
	if err != nil {
		return fmt.Errorf("insert product: %s %w", product.Name, err)
	}
	return nil
}

func (p *Postgres) InsertProducts(products []models.Product) (err error) {
	numProducts := len(products)
	index := 0
	numBatches := (numProducts + p.cfg.BatchSize - 1) / p.cfg.BatchSize
	args := make([]any, 0, 5*p.cfg.BatchSize)

	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback() // rollback после commit безопасен
	}(tx)

	_, err = tx.Exec("TRUNCATE TABLE public.product")
	if err != nil {
		return fmt.Errorf("truncate products: %w", err)
	}

	for batchIndex := 0; batchIndex < numBatches; batchIndex++ {
		curBatchSize := min(p.cfg.BatchSize, numProducts-batchIndex*p.cfg.BatchSize)
		query := newInsertProductQuery(curBatchSize)
		args = args[:0]
		for i := 0; i < curBatchSize; i++ {
			product := products[index]
			args = append(args, product.Name, product.Protein, product.Fat, product.Carbohydrate, product.Kcal)
			index++
		}

		_, err = tx.Exec(query, args...)
		if err != nil {
			return fmt.Errorf("batch insert products: %w", err)
		}
	}

	return tx.Commit()
}

func newInsertProductQuery(batchSize int) string {
	numberOfColumns := 5
	generatePlaceholders := func(row int) string {
		placeholders := make([]string, 0, numberOfColumns)
		for i := 0; i < numberOfColumns; i++ {
			placeholders = append(placeholders, fmt.Sprintf("$%d", row*numberOfColumns+i+1))
		}
		return "(" + strings.Join(placeholders, ",") + ")"
	}

	rows := make([]string, batchSize)
	for row := 0; row < batchSize; row++ {
		rows[row] = generatePlaceholders(row)
	}

	return "INSERT INTO product(name, protein, fat, carbohydrate, kcal) VALUES " + strings.Join(rows, ",")
}
