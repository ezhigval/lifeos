package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	financeapp "github.com/valentinezhov/lifeos/internal/finance/app"
	financedomain "github.com/valentinezhov/lifeos/internal/finance/domain"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type debtJSON struct {
	ID              string  `json:"id"`
	Creditor        string  `json:"creditor"`
	AmountCents     int64   `json:"amount_cents"`
	PaidCents       int64   `json:"paid_cents"`
	RemainingCents  int64   `json:"remaining_cents"`
	Currency        string  `json:"currency"`
	DueDate         *string `json:"due_date,omitempty"`
}

func debtToJSON(dto financeapp.DebtDTO) debtJSON {
	out := debtJSON{
		ID:             dto.ID.String(),
		Creditor:       dto.Creditor,
		AmountCents:    dto.AmountCents,
		PaidCents:      dto.PaidCents,
		RemainingCents: dto.RemainingCents,
		Currency:       dto.Currency,
	}
	if dto.DueDate != nil {
		s := dto.DueDate.Format("2006-01-02")
		out.DueDate = &s
	}
	return out
}

type transactionJSON struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency"`
	MonthTotal  int64  `json:"month_total"`
}

func transactionToJSON(dto financeapp.TransactionDTO) transactionJSON {
	return transactionJSON{
		ID:          dto.ID.String(),
		Description: dto.Description,
		AmountCents: dto.AmountCents,
		Currency:    dto.Currency,
		MonthTotal:  dto.MonthTotal,
	}
}

type recordIncomeRequest struct {
	AmountCents int64  `json:"amount_cents"`
	Description string `json:"description"`
	Currency    string `json:"currency"`
}

func (rt *Router) recordIncome(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req recordIncomeRequest
	if err := decodeJSON(r, &req); err != nil || req.AmountCents <= 0 {
		writeError(w, http.StatusBadRequest, "amount_cents is required")
		return
	}
	desc := strings.TrimSpace(req.Description)
	if desc == "" {
		desc = "доход"
	}
	dto, err := rt.deps.RecordIncome.Execute(r.Context(), financeapp.RecordIncomeInput{
		UserID:      userID,
		AmountCents: req.AmountCents,
		Currency:    req.Currency,
		Description: desc,
		Source:      events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, transactionToJSON(dto))
}

type recordExpenseRequest struct {
	AmountCents  int64  `json:"amount_cents"`
	CategoryName string `json:"category"`
	Currency     string `json:"currency"`
}

func (rt *Router) recordExpense(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req recordExpenseRequest
	if err := decodeJSON(r, &req); err != nil || req.AmountCents <= 0 {
		writeError(w, http.StatusBadRequest, "amount_cents is required")
		return
	}
	category := strings.TrimSpace(req.CategoryName)
	if category == "" {
		category = "Прочее"
	}
	dto, err := rt.deps.RecordExpense.Execute(r.Context(), financeapp.RecordExpenseInput{
		UserID:       userID,
		AmountCents:  req.AmountCents,
		Currency:     req.Currency,
		CategoryName: category,
		Source:       events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, transactionToJSON(dto))
}

func (rt *Router) cashFlow(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	summary, err := rt.deps.CashFlow.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"income_cents":  summary.IncomeCents,
		"expense_cents": summary.ExpenseCents,
		"net_cents":     summary.NetCents,
		"currency":      summary.Currency,
	})
}

func (rt *Router) listDebts(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := rt.deps.ListDebts.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]debtJSON, 0, len(items))
	for _, item := range items {
		out = append(out, debtToJSON(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"debts": out})
}

type createDebtRequest struct {
	Creditor    string  `json:"creditor"`
	AmountCents int64   `json:"amount_cents"`
	DueDate     *string `json:"due_date"`
}

func (rt *Router) createDebt(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req createDebtRequest
	if err := decodeJSON(r, &req); err != nil || strings.TrimSpace(req.Creditor) == "" || req.AmountCents <= 0 {
		writeError(w, http.StatusBadRequest, "creditor and amount_cents are required")
		return
	}
	var dueDate *time.Time
	if req.DueDate != nil && strings.TrimSpace(*req.DueDate) != "" {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(*req.DueDate))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid due_date, use YYYY-MM-DD")
			return
		}
		dueDate = &t
	}
	dto, err := rt.deps.CreateDebt.Execute(r.Context(), financeapp.CreateDebtInput{
		UserID:      userID,
		Creditor:    strings.TrimSpace(req.Creditor),
		AmountCents: req.AmountCents,
		DueDate:     dueDate,
		Source:      events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, debtToJSON(dto))
}

type payDebtRequest struct {
	AmountCents int64 `json:"amount_cents"`
}

func (rt *Router) payDebt(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	debtID, err := ids.ParseDebtID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid debt id")
		return
	}
	var req payDebtRequest
	if err := decodeJSON(r, &req); err != nil || req.AmountCents <= 0 {
		writeError(w, http.StatusBadRequest, "amount_cents is required")
		return
	}
	dto, err := rt.deps.PayDebt.Execute(r.Context(), financeapp.PayDebtInput{
		UserID:      userID,
		DebtID:      debtID,
		AmountCents: req.AmountCents,
		Source:      events.SourceHTTP,
	})
	if errors.Is(err, financedomain.ErrDebtNotFound) {
		writeError(w, http.StatusNotFound, "debt not found")
		return
	}
	if errors.Is(err, financedomain.ErrOverpayment) || errors.Is(err, financedomain.ErrDebtNotOpen) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, debtToJSON(dto))
}
