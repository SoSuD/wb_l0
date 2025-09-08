package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
	"wb_l0/config"
	"wb_l0/internal/kafka"
	"wb_l0/internal/orders/cache"
	"wb_l0/models"
	"wb_l0/pkg/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	gin    *gin.Engine
	port   string
	db     *pgxpool.Pool
	logger logger.Logger
	oKafka *kafka.Kafka
	oCache *cache.LRU[string, *models.Order]
	config config.Config
	wg     sync.WaitGroup
}

func NewServer(port string, db *pgxpool.Pool, logger logger.Logger, kafka *kafka.Kafka, config config.Config) *Server {
	return &Server{
		gin:    gin.Default(),
		port:   port,
		db:     db,
		logger: logger,
		oKafka: kafka,
		oCache: cache.New[string, *models.Order](config.Cache.Capacity, func(k string, v *models.Order) {
			fmt.Printf("evicted: %q -> %q\n", k, v)
		}),
		config: config,
	}
}

func (s *Server) Run(ctx context.Context) error {
	s.gin.Use(cors.Default())
	if err := s.Init(ctx, s.gin, s.config); err != nil {
		return err
	}

	// собираем http.Server (так можно мягко гасить)
	srv := &http.Server{
		Addr:              s.port,
		Handler:           s.gin,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	// слушаем в отдельной горутине
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	// ждём либо сигнал (ctx.Done), либо ошибку сервера
	select {
	case <-ctx.Done():
		// общий дедлайн
		to := 15 * time.Second
		if s.config.Server.ShutdownTimeout > 0 {
			to = time.Duration(s.config.Server.ShutdownTimeout) * time.Second // если у тебя значение в секундах
			// to = s.config.Server.ShutdownTimeout // если это уже time.Duration
		}
		shCtx, cancel := context.WithTimeout(context.Background(), to)
		defer cancel()

		// 1) мягко останавливаем HTTP — прекращаем приём, ждём активные
		if err := srv.Shutdown(shCtx); err != nil && !errors.Is(err, context.Canceled) {
			s.logger.Errorf("http shutdown: %v", err)
		}

		// 2) ждём консюмеров Kafka под тем же дедлайном
		done := make(chan struct{})
		go func() { s.wg.Wait(); close(done) }()
		select {
		case <-done:
			s.logger.Info("consumers stopped")
		case <-shCtx.Done():
			s.logger.Warn("shutdown deadline reached while waiting for consumers")
		}
		return nil
	case err := <-errCh:
		return err // сервер упал/закрылся
	}
}
