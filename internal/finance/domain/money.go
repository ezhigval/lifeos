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
	cents := m.AmountCents
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	rub := cents / 100
	kop := cents % 100
	if kop == 0 {
		return fmt.Sprintf("%s%d ₽", sign, rub)
	}
	return fmt.Sprintf("%s%d,%02d ₽", sign, rub, kop)
}
