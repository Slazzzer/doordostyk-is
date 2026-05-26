package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/doordostyk/api/internal/config"
	"github.com/doordostyk/api/internal/db"
	"github.com/doordostyk/api/internal/handler"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()
	log.Println("db: connected")

	if err := db.SeedPasswords(ctx, pool); err != nil {
		log.Printf("seed passwords: %v", err)
	}
	if err := db.EnsureSuppliers(ctx, pool); err != nil {
		log.Printf("seed suppliers: %v", err)
	}
	if err := db.EnsureExtraProducts(ctx, pool); err != nil {
		log.Printf("seed products: %v", err)
	}
	if err := db.EnsureStockReservationObjects(ctx, pool); err != nil {
		log.Printf("seed stock reservation: %v", err)
	}
	if err := db.EnsureNonNegativeBalances(ctx, pool); err != nil {
		log.Printf("seed balances: %v", err)
	}

	r := handler.NewRouter(pool, cfg)
	log.Printf("HTTP listening on :%s", cfg.HTTPPort)
	if err := r.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatalf("server: %v", err)
	}
}
