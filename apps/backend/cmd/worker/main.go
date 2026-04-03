package main

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	"github.com/verin/dms/apps/backend/internal/app"
	"github.com/verin/dms/apps/backend/internal/config"
	"github.com/verin/dms/apps/backend/internal/db"
	"github.com/verin/dms/apps/backend/internal/jobs"
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
	})
	if err != nil {
		panic(err)
	}

	redisOpt := asynq.RedisClientOpt{Addr: redisOptions.Addr, Password: redisOptions.Password, DB: redisOptions.DB}

	asynqServer := asynq.NewServer(redisOpt, asynq.Config{Concurrency: 10})
	asynqClient := asynq.NewClient(redisOpt)
	defer asynqClient.Close()

	server := app.NewServer(cfg, logger, pool, redisClient, storageClient, asynqClient)

	scheduler := asynq.NewScheduler(redisOpt, nil)
	retentionPayload, _ := json.Marshal(map[string]any{})
	_, _ = scheduler.Register("0 2 * * *", asynq.NewTask(jobs.TypeRetention, retentionPayload))
	_, _ = scheduler.Register("0 3 * * *", asynq.NewTask(jobs.TypeOrphanClean, retentionPayload))

	go func() {
		if err := scheduler.Run(); err != nil {
			logger.Error().Err(err).Msg("scheduler stopped")
		}
	}()

	logger.Info().Msg("starting worker with scheduler")
	if err := asynqServer.Run(server.WorkerMux()); err != nil {
		panic(err)
	}
}
