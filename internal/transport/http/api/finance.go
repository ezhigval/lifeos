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
	ID                  string  `json:"id"`
	Creditor            string  `json:"creditor"`
	AmountCents         int64   `json:"amount_cents"`
	PaidCents           int64   `json:"paid_cents"`
	RemainingCents      int64   `json:"remaining_cents"`
	Currency            string  `json:"currency"`
	DueDate             *string `json:"due_date,omitempty"`
	InstallmentCents    int64   `json:"installment_cents"`
	InstallmentInterval string  `json:"installment_interval"`
	NextPaymentDate     *string `json:"next_payment_date,omitempty"`
}

func debtToJSON(dto financeapp.DebtDTO) debtJSON {
	out := debtJSON{
		ID:                  dto.ID.String(),
		Creditor:            dto.Creditor,
		AmountCents:         dto.AmountCents,
		PaidCents:           dto.PaidCents,
		RemainingCents:      dto.RemainingCents,
		Currency:            dto.Currency,
		InstallmentCents:    dto.InstallmentCents,
		InstallmentInterval: dto.InstallmentInterval,
	}
	if dto.DueDate != nil {
		s := dto.DueDate.Format("2006-01-02")
		out.DueDate = &s
	}
	if dto.NextPaymentDate != nil {
		s := dto.NextPaymentDate.Format("2006-01-02")
		out.NextPaymentDate = &s
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

type financeCategoryJSON struct {
	Name        string  `json:"name"`
	AmountCents int64   `json:"amount_cents"`
	Percent     float64 `json:"percent"`
	ColorHint   *string `json:"color_hint,omitempty"`
}

type financeOverviewJSON struct {
	PeriodLabel  string                `json:"period_label"`
	IncomeCents  int64                 `json:"income_cents"`
	ExpenseCents int64                 `json:"expense_cents"`
	NetCents     int64                 `json:"net_cents"`
	Currency     string                `json:"currency"`
	Categories   []financeCategoryJSON `json:"categories"`
}

func (rt *Router) financeOverview(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.FinanceOverview == nil {
		writeError(w, http.StatusNotImplemented, "finance overview is not configured")
		return
	}
	period := strings.TrimSpace(r.URL.Query().Get("period"))
	dto, err := rt.deps.FinanceOverview.Execute(r.Context(), financeapp.FinanceOverviewInput{
		UserID: userID,
		Period: period,
	})
	if errors.Is(err, financeapp.ErrInvalidPeriod) {
		writeError(w, http.StatusBadRequest, "invalid period, use YYYY-MM")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	categories := make([]financeCategoryJSON, 0, len(dto.Categories))
	for _, c := range dto.Categories {
		categories = append(categories, financeCategoryJSON{
			Name:        c.Name,
			AmountCents: c.AmountCents,
			Percent:     c.Percent,
			ColorHint:   c.ColorHint,
		})
	}
	writeJSON(w, http.StatusOK, financeOverviewJSON{
		PeriodLabel:  dto.PeriodLabel,
		IncomeCents:  dto.IncomeCents,
		ExpenseCents: dto.ExpenseCents,
		NetCents:     dto.NetCents,
		Currency:     dto.Currency,
		Categories:   categories,
	})
}

func (rt *Router) listDebts(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.ListDebts == nil {
		writeError(w, http.StatusNotImplemented, "list debts is not configured")
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
	Creditor            string  `json:"creditor"`
	AmountCents         int64   `json:"amount_cents"`
	DueDate             *string `json:"due_date"`
	InstallmentCents    int64   `json:"installment_cents"`
	InstallmentInterval string  `json:"installment_interval"`
	NextPaymentDate     *string `json:"next_payment_date"`
}

func (rt *Router) createDebt(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.CreateDebt == nil {
		writeError(w, http.StatusNotImplemented, "create debt is not configured")
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
	var nextPay *time.Time
	if req.NextPaymentDate != nil && strings.TrimSpace(*req.NextPaymentDate) != "" {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(*req.NextPaymentDate))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid next_payment_date, use YYYY-MM-DD")
			return
		}
		nextPay = &t
	}
	dto, err := rt.deps.CreateDebt.Execute(r.Context(), financeapp.CreateDebtInput{
		UserID:              userID,
		Creditor:            strings.TrimSpace(req.Creditor),
		AmountCents:         req.AmountCents,
		DueDate:             dueDate,
		InstallmentCents:    req.InstallmentCents,
		InstallmentInterval: req.InstallmentInterval,
		NextPaymentDate:     nextPay,
		Source:              events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, debtToJSON(dto))
}

type payDebtRequest struct {
	AmountCents int64 `json:"amount_cents"`
	Regular     bool  `json:"regular"`
}

func (rt *Router) payDebt(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.PayDebt == nil {
		writeError(w, http.StatusNotImplemented, "pay debt is not configured")
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
		Regular:     req.Regular,
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

type planItemJSON struct {
	ID          string  `json:"id"`
	Kind        string  `json:"kind"`
	Title       string  `json:"title"`
	AmountCents int64   `json:"amount_cents"`
	Currency    string  `json:"currency"`
	Interval    string  `json:"interval"`
	NextDate    string  `json:"next_date"`
	Source      string  `json:"source"`
	DebtID      *string `json:"debt_id,omitempty"`
}

func planItemToJSON(dto financeapp.PlanItemDTO) planItemJSON {
	return planItemJSON{
		ID:          dto.ID,
		Kind:        dto.Kind,
		Title:       dto.Title,
		AmountCents: dto.AmountCents,
		Currency:    dto.Currency,
		Interval:    dto.Interval,
		NextDate:    dto.NextDate,
		Source:      dto.Source,
		DebtID:      dto.DebtID,
	}
}

func (rt *Router) listFinancePlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.ListFinancePlan == nil {
		writeError(w, http.StatusNotImplemented, "finance plan is not configured")
		return
	}
	plan, err := rt.deps.ListFinancePlan.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	items := make([]planItemJSON, 0, len(plan.Items))
	for _, item := range plan.Items {
		items = append(items, planItemToJSON(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":           items,
		"planned_income":  plan.PlannedIncome,
		"planned_expense": plan.PlannedExpense,
		"currency":        plan.Currency,
	})
}

type createPlannedRequest struct {
	Kind        string `json:"kind"`
	Title       string `json:"title"`
	AmountCents int64  `json:"amount_cents"`
	Interval    string `json:"interval"`
	NextDate    string `json:"next_date"`
}

func (rt *Router) createPlannedCashflow(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.CreatePlanned == nil {
		writeError(w, http.StatusNotImplemented, "create planned cashflow is not configured")
		return
	}
	var req createPlannedRequest
	if err := decodeJSON(r, &req); err != nil || strings.TrimSpace(req.Title) == "" || req.AmountCents <= 0 {
		writeError(w, http.StatusBadRequest, "kind, title and amount_cents are required")
		return
	}
	next := time.Now().UTC()
	if strings.TrimSpace(req.NextDate) != "" {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(req.NextDate))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid next_date, use YYYY-MM-DD")
			return
		}
		next = t
	}
	dto, err := rt.deps.CreatePlanned.Execute(r.Context(), financeapp.CreatePlannedCashflowInput{
		UserID:      userID,
		Kind:        strings.TrimSpace(req.Kind),
		Title:       strings.TrimSpace(req.Title),
		AmountCents: req.AmountCents,
		Interval:    strings.TrimSpace(req.Interval),
		NextDate:    next,
		Source:      events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, planItemToJSON(dto))
}

func (rt *Router) deletePlannedCashflow(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.DeletePlanned == nil {
		writeError(w, http.StatusNotImplemented, "delete planned cashflow is not configured")
		return
	}
	id, err := ids.ParsePlannedCashflowID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid plan id")
		return
	}
	if err := rt.deps.DeletePlanned.Execute(r.Context(), userID, id); err != nil {
		if errors.Is(err, financedomain.ErrPlanNotFound) {
			writeError(w, http.StatusNotFound, "planned cashflow not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (rt *Router) completePlannedCashflow(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.CompletePlanned == nil {
		writeError(w, http.StatusNotImplemented, "complete planned cashflow is not configured")
		return
	}
	id, err := ids.ParsePlannedCashflowID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid plan id")
		return
	}
	res, err := rt.deps.CompletePlanned.Execute(r.Context(), userID, id, events.SourceHTTP)
	if err != nil {
		if errors.Is(err, financedomain.ErrPlanNotFound) {
			writeError(w, http.StatusNotFound, "planned cashflow not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if res.Deleted {
		writeJSON(w, http.StatusOK, map[string]any{
			"deleted":       true,
			"posted":        res.Posted,
			"posted_cents":  res.PostedCents,
			"posted_kind":   res.PostedKind,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"deleted":      false,
		"item":         planItemToJSON(*res.Item),
		"posted":       res.Posted,
		"posted_cents": res.PostedCents,
		"posted_kind":  res.PostedKind,
	})
}
