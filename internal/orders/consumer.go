package orders

import "github.com/segmentio/kafka-go"

type Consumer interface {
	Create(msg *kafka.Message) error
}
