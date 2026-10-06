package main

import (
	"context"
	"log"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/seeder"
	"github.com/MamangRust/microservice-point-of-sale-pkg/dotenv"
	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
)

func main() {
	if err := dotenv.Viper(); err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	l, err := logger.NewLogger("seeder", nil)
	if err != nil {
		log.Fatalf("Error creating logger: %v", err)
	}

	// Each bounded context has its own PostgreSQL instance behind its own
	// PgBouncer, so the seeder opens one handle per context. Cross-context steps
	// (e.g. products→merchants) read through the owner's handle, which is why
	// every domain gets an explicit connection instead of one shared *sqlx.DB.
	identityDB, err := database.NewClientWithPrefix(l, database.IdentityCluster)
	if err != nil {
		log.Fatalf("Error connecting to identity database: %v", err)
	}
	defer identityDB.Close()

	merchantDB, err := database.NewClientWithPrefix(l, database.MerchantCluster)
	if err != nil {
		log.Fatalf("Error connecting to merchant database: %v", err)
	}
	defer merchantDB.Close()

	catalogDB, err := database.NewClientWithPrefix(l, database.CatalogCluster)
	if err != nil {
		log.Fatalf("Error connecting to catalog database: %v", err)
	}
	defer catalogDB.Close()

	salesDB, err := database.NewClientWithPrefix(l, database.SalesCluster)
	if err != nil {
		log.Fatalf("Error connecting to sales database: %v", err)
	}
	defer salesDB.Close()

	ctx := context.Background()

	s := seeder.NewSeeder(seeder.Deps{
		User:        identityDB,
		Role:        identityDB,
		Merchant:    merchantDB,
		Cashier:     merchantDB,
		Category:    catalogDB,
		Product:     catalogDB,
		Order:       salesDB,
		Transaction: salesDB,
		Ctx:         ctx,
		Logger:      l,
		Hash:        hash.NewHashingPassword(),
	})

	if err := s.Run(); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	l.Info("Seeding completed successfully.")
}
