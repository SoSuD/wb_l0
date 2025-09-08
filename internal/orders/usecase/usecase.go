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

func (u *OrdersUC) GetByID(ctx context.Context, orderID string) (*models.Order, error) {
	if u.cache != nil {
		if ord, ok := u.cache.Get(orderID); ok && ord != nil {
			u.logger.Infof("order found in cache: %s", orderID)
			return ord, nil
		}
		u.logger.Infof("order not found in cache: %s", orderID)
	} else {
		u.logger.Infof("cache is nil, skipping cache lookup")
	}
	ord, err := u.ordersRepo.GetByID(ctx, orderID)
	if err != nil {
		u.logger.Errorf("get order error: %v", err)
		return nil, err
	}
	if ord == nil {
		u.logger.Infof("order not found in repo: %s", orderID)
		return nil, nil
	}
	if u.cache != nil {
		u.cache.Put(ord.OrderUid, ord)
		u.logger.Infof("order cached: %s", ord.OrderUid)
	}

	return ord, nil
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
