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
	GetAllBudget(ctx context.Context) ([]Budget, error)
}

func NewUsecase(repo BudgetRepository) *Usecase {
	return &Usecase{
		repo: repo,
	}
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

	budgets, err := u.repo.GetAllBudget(ctx)
	if err != nil {
		return GetAllBudgetResponse{}, err
	}

	return GetAllBudgetResponse{
		budgets,
	}, nil
}
