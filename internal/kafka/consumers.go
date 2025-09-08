package kafka

import (
	"context"
	"errors"
	"time"

	"wb_l0/pkg/validation"
)

func (k *Kafka) Consume(ctx context.Context, f HandlerFunc) error {
	k.logger.Infof("kafka started: topic=%s group=%s", k.r.Config().Topic, k.r.Config().GroupID)
	defer func() {
		if err := k.r.Close(); err != nil {
			k.logger.Errorf("kafka reader close error: %v", err)
		}
		k.logger.Info("kafka stopped")
	}()

	for {
		msg, err := k.r.FetchMessage(ctx) // без автокоммита
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err() // shutdown
			}
			k.logger.Errorf("fetch message error: %v", err)
			time.Sleep(200 * time.Millisecond)
			continue
		}

		// ограничим обработку по времени
		hctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = f(hctx, &msg)
		cancel()

		if err != nil {
			// 1) перманентная (валидация) -> пропускаем (коммитим)
			var verr validation.ValidationErrors
			if errors.As(err, &verr) {
				k.logger.Errorf("validation error, skip: %v", verr)
				if cerr := k.r.CommitMessages(ctx, msg); cerr != nil {
					k.logger.Errorf("commit after validation error: %v", cerr)
				}
				continue
			}

			// 2) временная (таймаут/отмена/БД и т.п.) -> не коммитим, ретраим
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				k.logger.Warnf("handler canceled/timeout, will retry: %v", err)
			} else {
				k.logger.Errorf("transient handler error, will retry: %v", err)
			}
			time.Sleep(1 * time.Second)
			continue
		}

		// 3) успех -> коммит
		if err := k.r.CommitMessages(ctx, msg); err != nil {
			k.logger.Errorf("commit error: %v", err)
		}
	}
}
