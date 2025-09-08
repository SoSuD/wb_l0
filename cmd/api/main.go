package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"wb_l0/migrations"

	"wb_l0/config"
	"wb_l0/internal/kafka"
	"wb_l0/internal/server"
	"wb_l0/pkg/logger"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	cfgFile, err := config.LoadConfig("config/config-local.yml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err) // <— %v вместо %e
	}
	cfg, err := config.ParseConfig(cfgFile)
	if err != nil {
		log.Fatalf("failed to parse config: %v", err) // <— %v
	}

	appLogger := logger.NewApiLogger(cfg)
	appLogger.InitLogger()

	// Собираем DSN один раз
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.DbName,
		cfg.Postgres.SslMode,
	)

	// МИГРАЦИИ (вшитые в бинарь через go:embed)
	if err := applyMigrations(dsn, appLogger); err != nil {
		appLogger.Fatalf("migrations failed: %v", err)
	}
	appLogger.Infof("migrations applied")

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err) // <— %v
	}
	defer pool.Close()

	if err = pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping postgres: %v", err) // <— %v
	}
	appLogger.Infof("connected to postgres")
	appLogger.Infof("config: %+v", cfg)

	kaf := kafka.New(cfg.Kafka, pool, appLogger)
	s := server.NewServer(cfg.Server.Port, pool, appLogger, kaf, *cfg)

	if err := s.Run(ctx); err != nil {
		appLogger.Fatalf("s.Run: %v", err)
	}
	appLogger.Info("stopped process")
}

func applyMigrations(dsn string, logg logger.Logger) error {
	// Открываем подключение через database/sql с драйвером pgx
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("sql.Open: %w", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logg.Debug("failed to close db conn")
		}
	}(db)

	// Драйвер БД для migrate
	dbDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("postgres.WithInstance: %w", err)
	}

	// Источник миграций из embed FS
	src, err := iofs.New(migrations.Files, ".")
	if err != nil {
		return fmt.Errorf("iofs.New: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", dbDriver)
	if err != nil {
		return fmt.Errorf("migrate.NewWithInstance: %w", err)
	}
	defer func() {
		// Закрываем и чисто логируем ошибки источника/БД, если будут
		if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
			logg.Warnf("migrate.Close warnings: src=%v db=%v", srcErr, dbErr)
		}
	}()

	// Накатываем вверх; ErrNoChange — не ошибка
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate.Up: %w", err)
	}
	return nil
}
