package src

import (
	"context"
	"errors"
	"time"
)

type Usecase struct {
	repo BudgetRepository
}

type BudgetRepository interface {
	GetAllBudget() ([]Budget, error)
}

type Budget struct {
	ID          string
	Name        string
	Amount      float64
	Currency    string
	LastUpdated time.Time
}

type GetAllBudgetRequest struct{}
type GetAllBudgetResponse struct {
	Budgets []Budget
}

func (u *Usecase) GetAllBudget(ctx context.Context, req GetAllBudgetRequest) (GetAllBudgetResponse, error) {
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return GetAllBudgetResponse{}, errors.New("context cancelled")
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return GetAllBudgetResponse{}, errors.New("timed out")
		}
	}

	budgets, err := u.repo.GetAllBudget()
	if err != nil {
		return GetAllBudgetResponse{}, err
	}

	return GetAllBudgetResponse{
		budgets,
	}, nil
}
