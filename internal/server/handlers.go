package server

import (
	"context"
	"github.com/gin-gonic/gin"
	"wb_l0/config"
	ordersHttp "wb_l0/internal/orders/delivery/http"
	OrdersConsumer "wb_l0/internal/orders/delivery/kafka"
	ordersRepo "wb_l0/internal/orders/repository"
	ordersUC "wb_l0/internal/orders/usecase"
)

func (s *Server) Init(r *gin.Engine, config config.Config) error {

	oRepo := ordersRepo.NewOrdersRepo(s.db)
	oUC := ordersUC.NewOrdersUC(s.logger, oRepo, s.oCache)
	oUC.PutLastCache(context.Background(), config.Cache.Capacity)
	oH := ordersHttp.NewOrdersHandler(s.logger, oUC)
	ordersGroup := r.Group("/order")
	ordersHttp.MapOrdersRoutes(ordersGroup, oH)

	ohandler := OrdersConsumer.NewOrdersHandler(s.logger, oUC)
	OrdersConsumer.MapOrdersConsumers(s.oKafka, ohandler)

	return nil
}
