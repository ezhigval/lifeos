package domain

import (
	"errors"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var (
	ErrEmptyDescription = errors.New("transaction description is required")
	ErrInvalidKind      = errors.New("invalid transaction kind")
)

type Kind string

const (
	KindIncome  Kind = "income"
	KindExpense Kind = "expense"
)

func (k Kind) Valid() bool {
	return k == KindIncome || k == KindExpense
}

type Transaction struct {
	ID          ids.TransactionID
	UserID      ids.UserID
	CategoryID  ids.CategoryID
	Kind        Kind
	Money       Money
	Description string
	OccurredAt  time.Time
	CreatedAt   time.Time
}

func NewIncomeTransaction(
	userID ids.UserID,
	categoryID ids.CategoryID,
	amountCents int64,
	currency string,
	description string,
	occurredAt time.Time,
) (Transaction, error) {
	return newTransaction(userID, categoryID, KindIncome, amountCents, currency, description, occurredAt)
}

func NewExpenseTransaction(
	userID ids.UserID,
	categoryID ids.CategoryID,
	amountCents int64,
	currency string,
	description string,
	occurredAt time.Time,
) (Transaction, error) {
	return newTransaction(userID, categoryID, KindExpense, amountCents, currency, description, occurredAt)
}

func newTransaction(
	userID ids.UserID,
	categoryID ids.CategoryID,
	kind Kind,
	amountCents int64,
	currency string,
	description string,
	occurredAt time.Time,
) (Transaction, error) {
	if description == "" {
		return Transaction{}, ErrEmptyDescription
	}
	money, err := NewMoney(amountCents, currency)
	if err != nil {
		return Transaction{}, err
	}
	if !kind.Valid() {
		return Transaction{}, ErrInvalidKind
	}
	now := occurredAt.UTC()
	return Transaction{
		ID:          ids.NewTransactionID(),
		UserID:      userID,
		CategoryID:  categoryID,
		Kind:        kind,
		Money:       money,
		Description: description,
		OccurredAt:  now,
		CreatedAt:   now,
	}, nil
}
