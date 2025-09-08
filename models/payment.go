package models

import (
	"fmt"
	"wb_l0/pkg/validation"
)

type Payment struct {
	Transaction  string `json:"transaction" db:"transaction"`
	RequestId    string `json:"request_id" db:"request_id"`
	Currency     string `json:"currency" db:"currency"`
	Provider     string `json:"provider" db:"provider"`
	Amount       int    `json:"amount" db:"amount"`
	PaymentDt    int    `json:"payment_dt" db:"payment_dt"`
	Bank         string `json:"bank" db:"bank"`
	DeliveryCost int    `json:"delivery_cost" db:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total" db:"goods_total"`
	CustomFee    int    `json:"custom_fee" db:"custom_fee"`
}

func (p Payment) Validate() error {
	var errs validation.ValidationErrors

	if p.Transaction == "" {
		errs.Add("transaction", validation.ErrEmpty)
	}
	if p.Currency == "" {
		errs.Add("currency", validation.ErrEmpty)
	}
	if p.Provider == "" {
		errs.Add("provider", validation.ErrEmpty)
	}
	if p.Bank == "" {
		errs.Add("bank", validation.ErrEmpty)
	}

	if p.Amount < 0 {
		errs.Add("amount", validation.ErrNegative)
	}
	if p.DeliveryCost < 0 {
		errs.Add("delivery_cost", validation.ErrNegative)
	}
	if p.GoodsTotal < 0 {
		errs.Add("goods_total", validation.ErrNegative)
	}
	if p.CustomFee < 0 {
		errs.Add("custom_fee", validation.ErrNegative)
	}

	// Согласованность сумм
	expected := p.GoodsTotal + p.DeliveryCost + p.CustomFee
	if expected != 0 && p.Amount != expected {
		errs.Add("amount", fmt.Errorf("%w (expected sum=%d, got=%d)", validation.ErrInvalid, expected, p.Amount))
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
