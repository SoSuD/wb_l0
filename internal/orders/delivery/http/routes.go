package http

import (
	"github.com/gin-gonic/gin"
	"wb_l0/internal/orders"
)

func MapOrdersRoutes(ordersGroup *gin.RouterGroup, h orders.Handlers) {
	ordersGroup.GET("/:order_id", h.GetById())
}
