package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"ki.com/clean-arc-example/src"
)

type BudgetHTTP struct {
	uc BudgerUsecase
}

type BudgerUsecase interface {
	GetAllBudget(ctx context.Context, req src.GetAllBudgetRequest) (src.GetAllBudgetResponse, error)
}

type BudgetResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	LastUpdated time.Time `json:"last_updated"`
}

func NewBudgetHTTP(uc *src.Usecase) *BudgetHTTP {
	return &BudgetHTTP{uc: uc}
}

// GET /budgets
func (h *BudgetHTTP) GetAllBudgets(w http.ResponseWriter, r *http.Request) {
	// Only allow GET
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	// Optionally enforce a server-side timeout (tweak as needed)
	ctx := r.Context()
	// If you want a fixed timeout, uncomment:
	// var cancel context.CancelFunc
	// ctx, cancel = context.WithTimeout(r.Context(), 5*time.Second)
	// defer cancel()

	resp, err := h.uc.GetAllBudget(ctx, src.GetAllBudgetRequest{})
	if err != nil {
		// Map known errors to sensible HTTP status codes
		switch {
		case errors.Is(err, contextCanceledErr):
			// 499 is common in proxies for client-cancel; use 408 if you prefer standards
			w.WriteHeader(499)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "client cancelled request"})
			return
		case errors.Is(err, contextTimedOutErr):
			http.Error(w, "request timed out", http.StatusGatewayTimeout) // 504
			return
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	// Good practice for APIs: explicit caching policy (adjust as needed)
	w.Header().Set("Cache-Control", "no-store")
	// Optional: Add a response timestamp

	budgetResp := make([]BudgetResponse, 0, len(resp.Budgets))
	for _, b := range resp.Budgets {
		budgetResp = append(budgetResp, BudgetResponse{
			ID:          b.ID,
			Name:        b.Name,
			Amount:      b.Amount,
			Currency:    b.Currency,
			LastUpdated: b.LastUpdated,
		})
	}
	type out struct {
		Budgets   []BudgetResponse `json:"budgets"`
		Timestamp time.Time        `json:"timestamp"`
	}
	_ = json.NewEncoder(w).Encode(out{
		Budgets:   budgetResp,
		Timestamp: time.Now(),
	})
}

// Helpers to safely compare the usecase's sentinel errors.
// Your Usecase currently returns fresh errors.New(...), so we mirror them here.
var (
	contextCanceledErr = errors.New("context cancelled")
	contextTimedOutErr = errors.New("timed out")
)
