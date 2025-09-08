package validation

import (
	"errors"
	"strings"
)

// Сводная ошибка для удобного возврата сразу нескольких нарушений.
type FieldError struct {
	Path string // например: "order_uid", "delivery.email", "items[2].price"
	Err  error
}

type ValidationErrors []FieldError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	var b strings.Builder
	for i, fe := range ve {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(fe.Path)
		b.WriteString(": ")
		b.WriteString(fe.Err.Error())
	}
	return b.String()
}

func (ve *ValidationErrors) Add(path string, err error) {
	if err == nil {
		return
	}
	*ve = append(*ve, FieldError{Path: path, Err: err})
}

var (
	ErrEmpty       = errors.New("must not be empty")
	ErrInvalid     = errors.New("invalid value")
	ErrNotPositive = errors.New("must be > 0")
	ErrNegative    = errors.New("must be >= 0")
)
