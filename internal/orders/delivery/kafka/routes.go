package kafka

import (
	"context"
	"log"
	"wb_l0/internal/kafka"
)

func MapOrdersConsumers(kafka *kafka.Kafka, handler *OrdersHandler) {
	go func() {
		err := kafka.Consume(context.Background(), handler.Create)
		if err != nil {
			log.Fatal("Failed to run kafkaConsumer")
		}
	}()
}
