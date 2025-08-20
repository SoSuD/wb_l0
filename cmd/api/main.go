package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"wb_l0/config"
	"wb_l0/internal/kafka"
	"wb_l0/internal/server"
	"wb_l0/pkg/logger"
)

func main() {
	cfgFile, err := config.LoadConfig("config/config-local.yml")
	if err != nil {
		log.Fatalf("failed to load config %e", err)
	}
	cfg, err := config.ParseConfig(cfgFile)
	if err != nil {
		log.Fatalf("failed to parse config %e", err)
	}

	appLogger := logger.NewApiLogger(cfg)

	appLogger.InitLogger()
	pool, err := pgxpool.New(context.Background(), fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DbName, cfg.Postgres.SslMode))
	defer pool.Close()
	if err != nil {
		log.Fatalf("failed to connect to postgres %e", err)
	}
	if err = pool.Ping(context.Background()); err != nil {
		log.Fatalf("failed to ping postgres %e", err)
	}
	appLogger.Infof("connected to postgres")
	appLogger.Infof("config: %+v", cfg)
	kaf := kafka.New(cfg.Kafka, pool, appLogger)
	s := server.NewServer(cfg.Server.Port, pool, appLogger, kaf, cfg)

	if err := s.Run(); err != nil {
		appLogger.Fatalf("s.Run: %v", err)
	}

}
