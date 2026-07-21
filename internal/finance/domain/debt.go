package domain

import (
	"errors"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var (
	ErrEmptyCreditor     = errors.New("creditor is required")
	ErrInvalidDebtStatus = errors.New("invalid debt status")
	ErrDebtNotFound      = errors.New("debt not found")
	ErrDebtNotOpen       = errors.New("debt is not open")
	ErrOverpayment       = errors.New("payment exceeds remaining debt")
)

type DebtStatus string

const (
	DebtStatusOpen      DebtStatus = "open"
	DebtStatusPaid      DebtStatus = "paid"
	DebtStatusCancelled DebtStatus = "cancelled"
)

func (s DebtStatus) Valid() bool {
	return s == DebtStatusOpen || s == DebtStatusPaid || s == DebtStatusCancelled
}

type Debt struct {
	ID                  ids.DebtID
	UserID              ids.UserID
	Creditor            string
	AmountCents         int64
	PaidCents           int64
	DueDate             *time.Time
	InstallmentCents    int64
	InstallmentInterval string // none|weekly|monthly
	NextPaymentDate     *time.Time
	Status              DebtStatus
	CreatedAt           time.Time
}

func (d Debt) RemainingCents() int64 {
	return d.AmountCents - d.PaidCents
}

func NewDebt(userID ids.UserID, creditor string, amountCents int64, dueDate *time.Time, now time.Time) (Debt, error) {
	if creditor == "" {
		return Debt{}, ErrEmptyCreditor
	}
	if amountCents <= 0 {
		return Debt{}, ErrInvalidAmount
	}
	return Debt{
		ID:                  ids.NewDebtID(),
		UserID:              userID,
		Creditor:            creditor,
		AmountCents:         amountCents,
		PaidCents:           0,
		DueDate:             dueDate,
		InstallmentCents:    0,
		InstallmentInterval: "none",
		Status:              DebtStatusOpen,
		CreatedAt:           now.UTC(),
	}, nil
}

func (d *Debt) Pay(amountCents int64) error {
	if amountCents <= 0 {
		return ErrInvalidAmount
	}
	if d.Status != DebtStatusOpen {
		return ErrDebtNotOpen
	}
	if amountCents > d.RemainingCents() {
		return ErrOverpayment
	}
	d.PaidCents += amountCents
	if d.PaidCents >= d.AmountCents {
		d.Status = DebtStatusPaid
	}
	return nil
}

// AdvanceInstallment moves next_payment_date forward after a regular payment.
func (d *Debt) AdvanceInstallment(now time.Time) {
	if d.NextPaymentDate == nil || d.InstallmentInterval == "none" || d.InstallmentCents <= 0 {
		return
	}
	base := *d.NextPaymentDate
	switch d.InstallmentInterval {
	case "weekly":
		n := base.AddDate(0, 0, 7)
		d.NextPaymentDate = &n
	case "monthly":
		n := base.AddDate(0, 1, 0)
		d.NextPaymentDate = &n
	}
	_ = now
}
