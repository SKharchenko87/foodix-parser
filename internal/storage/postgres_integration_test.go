package storage

import (
	"database/sql"
	"os"
	"testing"

	"github.com/SKharchenko87/foodix-parser/internal/config"
	"github.com/SKharchenko87/foodix-parser/internal/models"
	_ "github.com/lib/pq" // postgres driver
)

func setupTestDB(t *testing.T) *sql.DB {
	os.Unsetenv("PGLOCALEDIR")
	dsn := "postgres://postgres:@localhost:5432/postgres?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to connect to test DB: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS product (
    name text,
    protein numeric(8,2),
    fat numeric(8,2),
    carbohydrate numeric(8,2),
    kcal int
);`)
	if err != nil {
		t.Fatalf("failed to create test table product: %v", err)
	}

	_, err = db.Exec(`TRUNCATE TABLE product`)
	if err != nil {
		t.Fatalf("failed to truncate test table product: %v", err)
	}

	return db
}

func TestInsertProduct(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	p := &Postgres{db: db, cfg: config.DB{BatchSize: 10}}
	product := models.Product{
		Name:         "Product Name",
		Protein:      2.0,
		Fat:          1.3,
		Carbohydrate: 1.4,
		Kcal:         150.0,
	}

	err := p.InsertProduct(product)
	if err != nil {
		t.Fatalf("failed to insert product: %v", err)
	}

	rows, err := db.Query("SELECT * FROM product WHERE name = $1", "Product Name")
	if err != nil {
		t.Fatalf("failed to query product: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var got models.Product
		if err := rows.Scan(&got.Name, &got.Protein, &got.Fat, &got.Carbohydrate, &got.Kcal); err != nil {
			t.Fatalf("failed to scan product: %v", err)
		}
		if got.Name != product.Name {
			t.Errorf("InsertProduct() = %v, want %v", got, product)
		}
	}
}

func TestInsertProducts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	p := &Postgres{db: db, cfg: config.DB{BatchSize: 10}}
	product0 := models.Product{
		Name:         "Product Name",
		Protein:      2.0,
		Fat:          1.3,
		Carbohydrate: 1.4,
		Kcal:         150.0,
	}

	product1 := models.Product{
		Name:         "Второй продукт",
		Protein:      1.0,
		Fat:          3.3,
		Carbohydrate: 4.4,
		Kcal:         250.0,
	}

	products := []models.Product{product0, product1}
	err := p.InsertProducts(products)
	if err != nil {
		t.Fatalf("failed to insert products: %v", err)
	}

	row := db.QueryRow("SELECT count(*) FROM product")
	var count int
	if err := row.Scan(&count); err != nil {
		t.Fatalf("failed to scan products: %v", err)
	}
	if count != len(products) {
		t.Errorf("InsertProducts() = %v, want %v", count, len(products))
	}
}
