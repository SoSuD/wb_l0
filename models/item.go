package models

import (
	"fmt"
	"wb_l0/pkg/validation"
)

type Item struct {
	ChrtId      int    `json:"chrt_id" db:"chrt_id"`
	TrackNumber string `json:"track_number" db:"track_number"`
	Price       int    `json:"price" db:"price"`
	Rid         string `json:"rid" db:"rid"`
	Name        string `json:"name" db:"name"`
	Sale        int    `json:"sale" db:"sale"`
	Size        string `json:"size" db:"size"`
	TotalPrice  int    `json:"total_price" db:"total_price"`
	NmId        int    `json:"nm_id" db:"nm_id"`
	Brand       string `json:"brand" db:"brand"`
	Status      int    `json:"status" db:"status"`
}

func (it Item) Validate() error {
	var errs validation.ValidationErrors

	if it.Name == "" {
		errs.Add("name", validation.ErrEmpty)
	}
	if it.Price <= 0 {
		errs.Add("price", validation.ErrNotPositive)
	}
	if it.TotalPrice < 0 {
		errs.Add("total_price", validation.ErrNegative)
	}
	if it.NmId <= 0 {
		errs.Add("nm_id", validation.ErrNotPositive)
	}
	if it.ChrtId <= 0 {
		errs.Add("chrt_id", validation.ErrNotPositive)
	}
	expected := it.Price
	if it.Sale > 0 {
		expected = int(float64(it.Price) * (1.0 - float64(it.Sale)/100.0))
	}
	if it.TotalPrice > 0 && it.TotalPrice != expected {
		errs.Add("total_price", fmt.Errorf("%w (expected=%d, got=%d)", validation.ErrInvalid, expected, it.TotalPrice))
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
