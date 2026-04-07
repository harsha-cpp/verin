package main

import (
	"context"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/verin/dms/apps/backend/internal/app"
	"github.com/verin/dms/apps/backend/internal/config"
	"github.com/verin/dms/apps/backend/internal/db"
	"github.com/verin/dms/apps/backend/internal/observability"
	"github.com/verin/dms/apps/backend/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := observability.NewLogger(cfg.AppEnv)
	ctx := context.Background()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	redisOptions, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		panic(err)
	}
	redisClient := redis.NewClient(redisOptions)
	defer redisClient.Close()

	endpoint, useSSL, err := storage.ParseEndpoint(cfg.S3Endpoint)
	if err != nil {
		panic(err)
	}
	storageClient, err := storage.New(storage.Config{
		Endpoint:     endpoint,
		AccessKey:    cfg.S3AccessKey,
		SecretKey:    cfg.S3SecretKey,
		UseSSL:       useSSL,
		UsePathStyle: cfg.S3UsePathStyle,
		Bucket:       cfg.S3Bucket,
		Region:       cfg.S3Region,
	})
	if err != nil {
		panic(err)
	}

	server := app.NewServer(cfg, logger, pool, redisClient, storageClient)
	server.StartBackgroundTasks(ctx)

	httpServer := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           server.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info().Str("addr", cfg.Addr()).Msg("starting api server")
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
