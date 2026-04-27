package main

import (
	"context"
	"log"
	"net/http"

	"github.com/beohoang98/moneyapp/internal/config"
	"github.com/beohoang98/moneyapp/internal/database"
	"github.com/beohoang98/moneyapp/internal/handlers"
	"github.com/beohoang98/moneyapp/internal/services"
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
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get underlying sql.DB: %v", err)
	}
	defer sqlDB.Close()

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

	authService := services.NewAuthService(db, cfg.JWTSecret, cfg.TokenExpiryHours)
	categoryService := services.NewCategoryService(db)
	expenseService := services.NewExpenseService(db, categoryService)
	incomeService := services.NewIncomeService(db, categoryService)
	invoiceService := services.NewInvoiceService(db)
	attachmentService := services.NewAttachmentService(db, store)
	dashboardService := services.NewDashboardService(db, invoiceService)

	expenseService.SetAttachmentService(attachmentService)
	incomeService.SetAttachmentService(attachmentService)
	invoiceService.SetAttachmentService(attachmentService)

	if _, err := invoiceService.UpdateOverdueStatuses(context.Background()); err != nil {
		log.Printf("Warning: failed to update overdue invoices on startup: %v", err)
	}

	publicMux := http.NewServeMux()
	protectedMux := http.NewServeMux()

	healthHandler := handlers.NewHealthHandler(db, store, cfg.StorageType)
	healthHandler.RegisterRoutes(publicMux)

	authHandler := handlers.NewAuthHandler(authService)
	authHandler.RegisterRoutes(publicMux)

	categoryHandler := handlers.NewCategoryHandler(categoryService)
	categoryHandler.RegisterRoutes(protectedMux)

	expenseHandler := handlers.NewExpenseHandler(expenseService)
	expenseHandler.RegisterRoutes(protectedMux)

	incomeHandler := handlers.NewIncomeHandler(incomeService)
	incomeHandler.RegisterRoutes(protectedMux)

	invoiceHandler := handlers.NewInvoiceHandler(invoiceService)
	invoiceHandler.RegisterRoutes(protectedMux)

	attachmentHandler := handlers.NewAttachmentHandler(attachmentService)
	attachmentHandler.RegisterRoutes(protectedMux)

	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	dashboardHandler.RegisterRoutes(protectedMux)

	exportHandler := handlers.NewExportHandler(expenseService, incomeService, categoryService)
	exportHandler.RegisterRoutes(protectedMux)

	scanningService := services.NewScanningService(db, store)
	scanningHandler := handlers.NewScanningHandler(scanningService)
	scanningHandler.RegisterRoutes(protectedMux)

	authMiddleware := handlers.AuthMiddleware(authService)

	mux := http.NewServeMux()
	mux.Handle("/api/health", publicMux)
	mux.Handle("/api/auth/", publicMux)
	mux.Handle("/api/", authMiddleware(protectedMux))

	handler := handlers.LoggingMiddleware(
		handlers.RecoveryMiddleware(
			handlers.CORSMiddleware(cfg.CORSAllowedOrigins)(mux),
		),
	)

	log.Printf("Server starting on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
