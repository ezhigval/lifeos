package domain

import (
	"errors"
	"fmt"
)

var ErrInvalidAmount = errors.New("amount must be positive")

type Money struct {
	AmountCents int64
	Currency    string
}

func NewMoney(amountCents int64, currency string) (Money, error) {
	if amountCents <= 0 {
		return Money{}, ErrInvalidAmount
	}
	if currency == "" {
		currency = "RUB"
	}
	return Money{AmountCents: amountCents, Currency: currency}, nil
}

func FormatMoney(m Money) string {
	rub := m.AmountCents / 100
	kop := m.AmountCents % 100
	if kop == 0 {
		return fmt.Sprintf("%d ₽", rub)
	}
	return fmt.Sprintf("%d,%02d ₽", rub, kop)
}
