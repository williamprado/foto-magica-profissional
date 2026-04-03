package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/williamprado/foto-magica-profissional/internal/config"
	"github.com/williamprado/foto-magica-profissional/internal/credits"
	"github.com/williamprado/foto-magica-profissional/internal/db"
	"github.com/williamprado/foto-magica-profissional/internal/generation"
	"github.com/williamprado/foto-magica-profissional/internal/logger"
	"github.com/williamprado/foto-magica-profissional/internal/notifications"
	"github.com/williamprado/foto-magica-profissional/internal/providers/ai"
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

	service := generation.NewService(
		pool,
		credits.NewService(pool),
		storageProvider,
		aiProvider,
		notifications.NewService(appLogger),
	)

	ticker := time.NewTicker(cfg.WorkerPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := service.ProcessNextQueuedJob(ctx, cfg.WorkerMaxAttempts); err != nil {
				appLogger.Error("worker_tick_failed", "error", err.Error())
			}
		}
	}
}
