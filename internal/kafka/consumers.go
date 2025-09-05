package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

func (k *Kafka) Consume(ctx context.Context, f HandlerFunc) error {
	k.logger.Infof("kafka kafka started: topic=%s group=%s", k.r.Config().Topic, k.r.Config().GroupID)
	defer func() {
		if err := k.r.Close(); err != nil {
			k.logger.Errorf("kafka reader close error: %v", err)
		}
		k.logger.Info("kafka kafka stopped")
	}()

	for {
		// ReadMessage делает fetch + автокоммит оффсета
		msg, err := k.r.ReadMessage(ctx)
		if err != nil {
			// если нас остановили через context — выходим
			if ctx.Err() != nil {
				return ctx.Err()
			}
			k.logger.Errorf("read message error: %v", err)
			time.Sleep(500 * time.Millisecond) // лёгкий бэкофф
			continue
		}
		go func(msg kafka.Message) {
			err = f(&msg)
			if err != nil {
				k.logger.Errorf("handler error: %v", err)
			}
		}(msg)
	}
}
