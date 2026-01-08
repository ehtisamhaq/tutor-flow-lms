package main

import (
	"os"

	"go.uber.org/zap"

	"github.com/tutorflow/tutorflow-server/internal/app"
	"github.com/tutorflow/tutorflow-server/internal/pkg/config"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	if os.Getenv("APP_ENV") == "development" {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		sugar.Fatalf("Failed to load config: %v", err)
	}

	// Initialize and run the application
	application := app.New(cfg, sugar)
	if err := application.Run(); err != nil {
		sugar.Fatalf("Application failed: %v", err)
	}
}
