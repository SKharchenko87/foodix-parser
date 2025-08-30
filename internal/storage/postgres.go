package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/SKharchenko87/foodix-parser/internal/config"
	"github.com/SKharchenko87/foodix-parser/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool     *pgxpool.Pool
	cfg      config.DB
	user     string
	password string
	db       string
	host     string
	port     int
}

func NewPostgres(cfg config.DB) (DB, error) {
	res := new(Postgres)
	res.cfg = cfg
	err := res.readPostgresEnv()
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.New(context.Background(), res.generateConnectionString())
	if err != nil {
		return nil, err
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	res.pool = pool

	return res, nil
}

func (p *Postgres) readPostgresEnv() error {
	var err error
	user, exists := os.LookupEnv("POSTGRES_USER")
	if !exists {
		err = errors.Join(err, errors.New("POSTGRES_USER not found"))
	}
	password, exists := os.LookupEnv("POSTGRES_PASSWORD")
	if !exists {
		err = errors.Join(err, errors.New("POSTGRES_PASSWORD not found"))
	}
	db, exists := os.LookupEnv("POSTGRES_DB")
	if !exists {
		err = errors.Join(err, errors.New("POSTGRES_DB not found"))
	}
	host, exists := os.LookupEnv("POSTGRES_HOST")
	if !exists {
		err = errors.Join(err, errors.New("POSTGRES_HOST not found"))
	}
	portStr, exists := os.LookupEnv("POSTGRES_PORT")
	if !exists {
		err = errors.Join(err, errors.New("POSTGRES_PORT not found"))
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		err = errors.Join(err, errors.New("POSTGRES_PORT must be an integer"))
	}

	p.user = user
	p.password = password
	p.host = host
	p.port = port
	p.db = db

	return err
}

func (p Postgres) generateConnectionString() string {
	return fmt.Sprintf("%s://%s:%s@%s:%d/%s", p.cfg.Name, p.user, p.password, p.host, p.port, p.db)
}

func (p *Postgres) Close() error {
	p.pool.Close()
	return nil
}

func (p *Postgres) InsertProduct(product models.Product) (err error) {
	_, err = p.pool.Exec(context.Background(),
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

	ctx := context.Background()
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func(tx pgx.Tx) {
		_ = tx.Rollback(ctx) // rollback после commit безопасен
	}(tx)

	_, err = tx.Exec(ctx, "TRUNCATE TABLE public.product")
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

		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("batch insert products: %w", err)
		}
	}

	return tx.Commit(ctx)
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
