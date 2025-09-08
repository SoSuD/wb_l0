package models

import (
	"fmt"
	"strings"
	"time"
	"wb_l0/pkg/validation"
)

type Order struct {
	OrderUid          string    `json:"order_uid" db:"order_uid"`
	TrackNumber       string    `json:"track_number" db:"track_number"`
	Entry             string    `json:"entry" db:"entry"`
	Delivery          Delivery  `json:"delivery" db:"delivery"`
	Payment           Payment   `json:"payment" db:"payment"`
	Items             []Item    `json:"items" db:"items"`
	Locale            string    `json:"locale" db:"locale"`
	InternalSignature string    `json:"internal_signature" db:"internal_signature"`
	CustomerId        string    `json:"customer_id" db:"customer_id"`
	DeliveryService   string    `json:"delivery_service" db:"delivery_service"`
	Shardkey          string    `json:"shardkey" db:"shardkey"`
	SmId              int       `json:"sm_id" db:"sm_id"`
	DateCreated       time.Time `json:"date_created" db:"date_created"`
	OofShard          string    `json:"oof_shard" db:"oof_shard"`
}

func (o Order) Validate() error {
	var errs validation.ValidationErrors

	// Базовые обязательные поля
	if strings.TrimSpace(o.OrderUid) == "" {
		errs.Add("order_uid", validation.ErrEmpty)
	}
	if strings.TrimSpace(o.TrackNumber) == "" {
		errs.Add("track_number", validation.ErrEmpty)
	}
	if strings.TrimSpace(o.Entry) == "" {
		errs.Add("entry", validation.ErrEmpty)
	}
	if strings.TrimSpace(o.CustomerId) == "" {
		errs.Add("customer_id", validation.ErrEmpty)
	}
	if strings.TrimSpace(o.DeliveryService) == "" {
		errs.Add("delivery_service", validation.ErrEmpty)
	}
	if strings.TrimSpace(o.Shardkey) == "" {
		errs.Add("shardkey", validation.ErrEmpty)
	}
	if strings.TrimSpace(o.OofShard) == "" {
		errs.Add("oof_shard", validation.ErrEmpty)
	}
	// Коллекции
	if len(o.Items) == 0 {
		errs.Add("items", validation.ErrEmpty)
	}

	// Вложенные объекты
	if derr := o.Delivery.Validate(); derr != nil {
		errs.Add("delivery", derr)
	}
	if perr := o.Payment.Validate(); perr != nil {
		errs.Add("payment", perr)
	}

	// Валидация каждого Item + согласованность с заказом
	for i, it := range o.Items {
		if ierr := it.Validate(); ierr != nil {
			errs.Add(fmt.Sprintf("items[%d]", i), ierr)
		}
		// Бизнес-правило: track_number в item должен совпадать с заказом (часто встречается)
		if strings.TrimSpace(it.TrackNumber) != "" && it.TrackNumber != o.TrackNumber {
			errs.Add(
				fmt.Sprintf("items[%d].track_number", i),
				fmt.Errorf("%w (must equal order.track_number)", validation.ErrInvalid))
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
