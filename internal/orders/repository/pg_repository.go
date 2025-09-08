package repository

import (
	"context"
	"wb_l0/internal/orders"
	"wb_l0/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ordersRepo struct {
	db *pgxpool.Pool
}

func NewOrdersRepo(db *pgxpool.Pool) orders.Repository {
	return &ordersRepo{db: db}
}

func (r *ordersRepo) Create(ctx context.Context, order *models.Order) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)

	}()

	_, err = tx.Exec(ctx, insertOrder, order.OrderUid, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerId, order.DeliveryService, order.Shardkey, order.SmId, order.DateCreated, order.OofShard)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, insertDelivery, order.OrderUid, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, insertPayment, order.OrderUid, order.Payment.Transaction, order.Payment.RequestId, order.Payment.Currency, order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return err
	}
	for _, item := range order.Items {
		_, err = tx.Exec(ctx, insertItem, order.OrderUid, item.ChrtId, item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmId, item.Brand, item.Status)
		if err != nil {
			return err
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil

}

func (r *ordersRepo) GetByID(ctx context.Context, orderId string) (*models.Order, error) {
	var order models.Order

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})

	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	err = pgxscan.Get(ctx, tx, &order, getOrder, orderId)
	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	err = pgxscan.Get(ctx, tx, &order.Payment, getOrderPayments, orderId)
	if err != nil {
		return nil, err
	}

	err = pgxscan.Get(ctx, tx, &order.Delivery, getOrderDeliveries, orderId)
	if err != nil {
		return nil, err
	}
	err = pgxscan.Select(ctx, tx, &order.Items, getOrderItems, orderId)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *ordersRepo) GetLastByCount(ctx context.Context, count int) ([]*models.Order, error) {
	var ord []*models.Order
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	err = pgxscan.Select(ctx, tx, &ord, getLastOrders, count)
	if err != nil {
		return nil, err
	}
	for _, o := range ord {
		err = pgxscan.Get(ctx, tx, &o.Payment, getOrderPayments, o.OrderUid)
		if err != nil {
			return nil, err
		}
		err = pgxscan.Get(ctx, tx, &o.Delivery, getOrderDeliveries, o.OrderUid)
		if err != nil {
			return nil, err
		}
		err = pgxscan.Select(ctx, tx, &o.Items, getOrderItems, o.OrderUid)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}
	return ord, nil

}
