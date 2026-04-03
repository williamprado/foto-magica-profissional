package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/williamprado/foto-magica-profissional/internal/config"
	"github.com/williamprado/foto-magica-profissional/internal/db"
	httpx "github.com/williamprado/foto-magica-profissional/internal/http"
	"github.com/williamprado/foto-magica-profissional/internal/logger"
	"github.com/williamprado/foto-magica-profissional/internal/notifications"
	"github.com/williamprado/foto-magica-profissional/internal/providers/ai"
	"github.com/williamprado/foto-magica-profissional/internal/providers/payment"
	"github.com/williamprado/foto-magica-profissional/internal/providers/storage"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	appLogger := logger.New(cfg.AppEnv)
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, pool, cfg.AppBaseDir); err != nil {
		log.Fatal(err)
	}

	storageProvider, err := storage.NewLocalProvider(cfg.LocalStoragePath)
	if cfg.StorageDriver != "local" {
		storageProvider, err = storage.NewS3Provider(ctx, cfg.StorageBucket, cfg.StorageRegion, cfg.StorageEndpoint, cfg.StorageAccessKey, cfg.StorageSecretKey, cfg.StorageUsePathStyle)
	}
	if err != nil {
		log.Fatal(err)
	}

	var aiProvider ai.Provider = ai.MockProvider{}
	if cfg.GoogleAPIKey != "" {
		aiProvider, err = ai.NewGoogleProvider(ctx, cfg.GoogleAPIKey, cfg.GoogleAnalysisModel, cfg.GooglePromptModel, cfg.GoogleImageModel)
		if err != nil {
			log.Fatal(err)
		}
	}

	router := httpx.NewRouter(httpx.Deps{
		Config:  cfg,
		Logger:  appLogger,
		DB:      pool,
		Storage: storageProvider,
		AI:      aiProvider,
		Payment: map[string]payment.Provider{
			"mock": payment.MockProvider{},
		},
		Notifications: notifications.NewService(appLogger),
	})

	if err := httpx.Serve(ctx, cfg, router); err != nil && err.Error() != "http: Server closed" {
		log.Fatal(err)
	}
}
