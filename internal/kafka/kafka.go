package kafka

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
	"wb_l0/config"
	"wb_l0/pkg/logger"
)

type Kafka struct {
	r      *kafka.Reader
	db     *pgxpool.Pool
	logger logger.Logger
}

type HandlerFunc func(msg *kafka.Message) error

func New(config config.Kafka, db *pgxpool.Pool, logger logger.Logger) *Kafka {
	return &Kafka{
		r: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  config.Orders.Brokers,
			Topic:    config.Orders.Topic,
			GroupID:  config.Orders.GroupId,
			MinBytes: 1,
			MaxBytes: 10e6,
			MaxWait:  1500 * time.Millisecond,
			// Удобный логгер
			//Logger:      kafka.LoggerFunc(func(msg string, _ ...interface{}) { log.Println("[kafka]", msg) }),
			ErrorLogger: kafka.LoggerFunc(func(msg string, _ ...interface{}) { log.Println("[kafka-err]", msg) }),
		}),
		db:     db,
		logger: logger,
	}
}
