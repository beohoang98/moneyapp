package main

import (
	"log"
	"net/http"

	"github.com/beohoang98/moneyapp/internal/config"
	"github.com/beohoang98/moneyapp/internal/database"
	"github.com/beohoang98/moneyapp/internal/handlers"
	"github.com/beohoang98/moneyapp/internal/storage"
	"github.com/beohoang98/moneyapp/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.Open(cfg.DBPath, migrations.FS)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	var store storage.ObjectStore
	switch cfg.StorageType {
	case "local":
		store, err = storage.NewLocalStorage(cfg.LocalStoragePath)
		if err != nil {
			log.Fatalf("failed to init local storage at %q: %v", cfg.LocalStoragePath, err)
		}
		log.Printf("Storage: local (%s)", cfg.LocalStoragePath)
	case "s3":
		store, err = storage.NewMinIOStorage(
			cfg.MinIOEndpoint,
			cfg.MinIOAccessKey,
			cfg.MinIOSecretKey,
			cfg.MinIOBucket,
			cfg.MinIOUseSSL,
		)
		if err != nil {
			log.Fatalf("failed to init S3 storage: %v", err)
		}
		log.Print("Storage: s3")
	}

	mux := http.NewServeMux()

	healthHandler := handlers.NewHealthHandler(db, store, cfg.StorageType)
	healthHandler.RegisterRoutes(mux)

	frontendOrigin := "http://localhost:5173"
	handler := handlers.LoggingMiddleware(
		handlers.RecoveryMiddleware(
			handlers.CORSMiddleware(frontendOrigin)(mux),
		),
	)

	log.Printf("Server starting on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
