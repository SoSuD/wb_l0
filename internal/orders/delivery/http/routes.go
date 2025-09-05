package http

import (
	"wb_l0/internal/orders"

	"github.com/gin-gonic/gin"
)

func MapOrdersRoutes(ordersGroup *gin.RouterGroup, h orders.Handlers) {
	ordersGroup.GET("/:order_id", h.GetById())
}
