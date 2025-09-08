package kafka

import (
	"context"
	"errors"
	"log"
	"sync"
	"wb_l0/internal/kafka"
)

func StartConsumers(ctx context.Context, wg *sync.WaitGroup, kafka *kafka.Kafka, handler *OrdersHandler) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := kafka.Consume(ctx, handler.Create)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Println("Kafka consumer terminated by ctx")
			} else {
				log.Printf("Failed to run kafkaConsumer: %e", err)
			}
		}
	}()
}
