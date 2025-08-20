package kafka

import (
	"context"
	"wb_l0/internal/kafka"
)

func MapOrdersConsumers(kafka *kafka.Kafka, handler *OrdersHandler) {
	go kafka.Consume(context.Background(), handler.Create)
}
