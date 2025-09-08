package orders

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Consumer interface {
	Create(ctx context.Context, msg *kafka.Message) error
}
