package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"golang.org/x/exp/rand"
)

const seedCreateProductSQL = `
INSERT INTO products (merchant_id, category_id, name, description, price, count_in_stock, brand, weight, slug_product, image_product, barcode)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING product_id`

const seedGetProductsByMerchantSQL = `
SELECT product_id, price
FROM products p
JOIN categories c ON p.category_id = c.category_id
WHERE p.deleted_at IS NULL
  AND p.merchant_id = $1
  AND (
      p.name ILIKE '%' || COALESCE($2, '') || '%'
      OR p.description ILIKE '%' || COALESCE($2, '') || '%'
      OR $2 IS NULL
  )
  AND (
      c.category_id = NULLIF($3, 0)
      OR NULLIF($3, 0) IS NULL
  )
  AND (
      p.price >= COALESCE(NULLIF($4, 0), 0)
      AND p.price <= COALESCE(NULLIF($5, 0), 999999999)
  )
ORDER BY p.created_at DESC
LIMIT $6 OFFSET $7`

type seedProductRow struct {
	ProductID int32 `db:"product_id"`
	Price     int32 `db:"price"`
}

type productSeeder struct {
	db         *sqlx.DB
	categoryDB *sqlx.DB
	merchantDB *sqlx.DB
	ctx        context.Context
	logger     logger.LoggerInterface
}

func NewProductSeeder(db *sqlx.DB, categoryDB *sqlx.DB, merchantDB *sqlx.DB, ctx context.Context, logger logger.LoggerInterface) *productSeeder {
	return &productSeeder{db: db, categoryDB: categoryDB, merchantDB: merchantDB, ctx: ctx, logger: logger}
}

func (r *productSeeder) Seed() error {
	var merchants []seedMerchantRow
	if err := r.merchantDB.SelectContext(r.ctx, &merchants, seedGetMerchantsSQL, "", 20, 0); err != nil {
		r.logger.Error("Failed to get merchants:", zap.Any("error", err))
		return err
	}

	var categories []seedCategoryRow
	if err := r.categoryDB.SelectContext(r.ctx, &categories, seedGetCategoriesSQL, "", 20, 0); err != nil {
		r.logger.Error("Failed to get categories:", zap.Any("error", err))
		return err
	}

	if len(merchants) == 0 || len(categories) == 0 {
		r.logger.Error("No merchants or categories found. Product seeding aborted.")
		return fmt.Errorf("no merchants or categories found for product seeding")
	}

	productNames := []string{
		"Smartphone", "Laptop", "Wireless Earbuds", "Gaming Mouse", "Mechanical Keyboard",
		"Smartwatch", "Power Bank", "Bluetooth Speaker", "External Hard Drive", "Monitor",
	}
	brands := []string{"Samsung", "Apple", "Sony", "Logitech", "Razer", "Xiaomi", "HP", "Dell", "Acer", "Asus"}
	images := []string{
		"image1.jpg", "image2.jpg", "image3.jpg", "image4.jpg", "image5.jpg",
		"image6.jpg", "image7.jpg", "image8.jpg", "image9.jpg", "image10.jpg",
	}

	rand.Seed(uint64(time.Now().UnixNano()))

	for i := 0; i < 10; i++ {
		merchant := merchants[rand.Intn(len(merchants))]
		category := categories[rand.Intn(len(categories))]
		name := productNames[rand.Intn(len(productNames))]
		brand := brands[rand.Intn(len(brands))]
		price := int32(rand.Intn(5000000) + 50000)
		countInStock := int32(rand.Intn(100) + 1)
		weight := int32(rand.Intn(5000) + 100)
		slug := fmt.Sprintf("%s-%d", name, rand.Intn(1000))
		image := images[rand.Intn(len(images))]
		barcode := fmt.Sprintf("BC-%d", rand.Intn(9999999))

		var created seedProductRow
		if err := r.db.GetContext(r.ctx, &created, seedCreateProductSQL,
			merchant.MerchantID, category.CategoryID, name,
			ptrString(fmt.Sprintf("Description for %s", name)), price, countInStock,
			ptrString(brand), ptrInt32(weight), ptrString(slug), ptrString(image), ptrString(barcode)); err != nil {
			r.logger.Error("Failed to create product:", zap.Error(err))
			return err
		}
	}

	r.logger.Info("Product seeding completed successfully.", zap.Int("count", 10))
	return nil
}
