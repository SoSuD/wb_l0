package server

import (
	"fmt"
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
}

func NewServer(port string, db *pgxpool.Pool, logger logger.Logger, kafka *kafka.Kafka, config *config.Config) *Server {
	return &Server{
		gin:    gin.Default(),
		port:   port,
		db:     db,
		logger: logger,
		oKafka: kafka,
		oCache: cache.New[string, *models.Order](config.Cache.Capacity, func(k string, v *models.Order) {
			fmt.Printf("evicted: %q -> %q\n", k, v)
		}),
	}
}

func (s *Server) Run() error {
	s.gin.Use(cors.Default())

	if err := s.Init(s.gin, s.config); err != nil {
		return err
	}

	return s.gin.Run(s.port)
}
