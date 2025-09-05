package http

import (
	"wb_l0/internal/orders"
	"wb_l0/models"
	"wb_l0/pkg/logger"

	"github.com/gin-gonic/gin"
)

type OrdersHandler struct {
	logger logger.Logger
	oUC    orders.UseCase
}

func NewOrdersHandler(logger logger.Logger, oUC orders.UseCase) *OrdersHandler {
	return &OrdersHandler{
		logger: logger,
		oUC:    oUC,
	}
}

func (h *OrdersHandler) GetById() gin.HandlerFunc {
	type request struct {
		OrderId string `uri:"order_id" binding:"required"`
	}
	type response struct {
		models.Order
	}
	return func(c *gin.Context) {
		var req request
		if err := c.ShouldBindUri(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		order, err := h.oUC.GetByID(c, req.OrderId)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, response{*order})
		return
	}

}
