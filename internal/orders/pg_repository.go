package orders

import (
	"context"
	"wb_l0/models"
)

type Repository interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, orderId string) (*models.Order, error)
	GetLastByCount(ctx context.Context, count int) ([]*models.Order, error)
}
