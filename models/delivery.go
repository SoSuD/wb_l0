package models

import (
	"strings"
	"wb_l0/pkg/validation"
)

type Delivery struct {
	Name    string `json:"name" db:"name"`
	Phone   string `json:"phone" db:"phone"`
	Zip     string `json:"zip" db:"zip"`
	City    string `json:"city" db:"city"`
	Address string `json:"address" db:"address"`
	Region  string `json:"region" db:"region"`
	Email   string `json:"email" db:"email"`
}

func (d Delivery) Validate() error {
	var errs validation.ValidationErrors

	if strings.TrimSpace(d.Name) == "" {
		errs.Add("name", validation.ErrEmpty)
	}
	if strings.TrimSpace(d.Phone) == "" {
		errs.Add("phone", validation.ErrEmpty)
	}
	if strings.TrimSpace(d.City) == "" {
		errs.Add("city", validation.ErrEmpty)
	}
	if strings.TrimSpace(d.Address) == "" {
		errs.Add("Address", validation.ErrEmpty)
	}
	if strings.TrimSpace(d.Email) == "" {
		errs.Add("email", validation.ErrEmpty)
	} else if !strings.Contains(d.Email, "@") {
		errs.Add("email", validation.ErrInvalid)
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
