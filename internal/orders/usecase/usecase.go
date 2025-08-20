package usecase

import (
	"context"
	"wb_l0/internal/orders"
	"wb_l0/internal/orders/cache"
	"wb_l0/models"
	"wb_l0/pkg/logger"
)

type OrdersUC struct {
	logger     logger.Logger
	ordersRepo orders.Repository
	cache      *cache.LRU[string, *models.Order]
}

func NewOrdersUC(logger logger.Logger, ordersRepo orders.Repository, cache *cache.LRU[string, *models.Order]) *OrdersUC {
	return &OrdersUC{
		logger:     logger,
		ordersRepo: ordersRepo,
		cache:      cache,
	}
}

func (u *OrdersUC) Create(ctx context.Context, order *models.Order) error {
	if err := u.ordersRepo.Create(ctx, order); err != nil {
		u.logger.Errorf("create order error: %v", err)
		return err
	}
	u.logger.Infof("order created: %v", order)
	u.cache.Put(order.OrderUid, order)
	u.logger.Infof("order cached: %v", order)
	return nil

}

func (u *OrdersUC) GetByID(ctx context.Context, orderId string) (*models.Order, error) {
	if order, ok := u.cache.Get(orderId); ok {
		u.logger.Infof("order found in cache: %v", order)
		return order, nil
	} else {
		u.logger.Infof("order not found in cache: %v", orderId)
	}

	if order, err := u.ordersRepo.GetByID(ctx, orderId); err != nil {
		u.logger.Errorf("get order error: %v", err)
		return nil, err
	} else {
		u.logger.Infof("order found: %v", order)
		u.cache.Put(order.OrderUid, order)
		u.logger.Infof("order cached: %v", order)
		return order, nil
	}
}

func (u *OrdersUC) PutLastCache(ctx context.Context, count int) error {
	ords, err := u.ordersRepo.GetLastByCount(ctx, count)
	if err != nil {
		u.logger.Errorf("get last orders error: %v", err)
		return err
	}
	for _, o := range ords {
		u.cache.Put(o.OrderUid, o)
	}
	return nil
}
