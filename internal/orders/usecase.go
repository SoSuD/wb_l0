package orders

import (
	"context"
	"wb_l0/models"
)

type UseCase interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, orderId string) (*models.Order, error)
	PutLastCache(ctx context.Context, count int) error
}
