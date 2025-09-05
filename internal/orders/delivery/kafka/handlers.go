package kafka

import (
	"context"
	"encoding/json"
	"wb_l0/internal/orders"
	"wb_l0/models"
	"wb_l0/pkg/logger"

	"github.com/segmentio/kafka-go"
)

type OrdersHandler struct {
	oUC    orders.UseCase
	logger logger.Logger
}

func NewOrdersHandler(logger logger.Logger, oUC orders.UseCase) *OrdersHandler {
	return &OrdersHandler{
		oUC:    oUC,
		logger: logger,
	}
}

func (h *OrdersHandler) Create(msg *kafka.Message) error {
	var order models.Order
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		h.logger.Errorf("unmarshal order error: %v", err)
		return err
	}

	if err := h.oUC.Create(context.Background(), &order); err != nil {
		h.logger.Errorf("create order error: %v", err)
		return err
	}
	return nil
}
