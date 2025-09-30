package grpc

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"ki.com/clean-arc-example/src"
	budgetv1 "ki.com/clean-arc-example/src/grpc/example.com/yourmod/api/budget/v1"
)

type BudgetServer struct {
	budgetv1.UnimplementedBudgetServiceServer
	uc BudgerUsecase
}

type BudgerUsecase interface {
	GetAllBudget(ctx context.Context, req src.GetAllBudgetRequest) (src.GetAllBudgetResponse, error)
}

func NewBudgetServer(uc *src.Usecase) *BudgetServer {
	return &BudgetServer{uc: uc}
}

func (s *BudgetServer) ListBudgets(ctx context.Context, in *budgetv1.ListBudgetsRequest) (*budgetv1.ListBudgetsResponse, error) {
	// Respect client-set deadlines and cancellations
	if err := ctx.Err(); err != nil {
		return nil, toStatus(err)
	}

	// Call usecase
	resp, err := s.uc.GetAllBudget(ctx, src.GetAllBudgetRequest{})
	if err != nil {
		// Map domain or sentinel errors to gRPC codes
		return nil, toUsecaseStatus(err)
	}

	// Map domain -> protobuf
	out := &budgetv1.ListBudgetsResponse{
		Budgets:    make([]*budgetv1.Budget, 0, len(resp.Budgets)),
		ServerTime: timestamppb.New(time.Now()),
	}
	for _, b := range resp.Budgets {
		out.Budgets = append(out.Budgets, &budgetv1.Budget{
			Id:          b.ID,
			Name:        b.Name,
			Amount:      b.Amount,
			Currency:    b.Currency,
			LastUpdated: timestamppb.New(b.LastUpdated),
		})
	}

	return out, nil
}

// Translate context errors to gRPC status codes.
func toStatus(err error) error {
	switch err {
	case context.Canceled:
		return status.Error(codes.Canceled, "client canceled request")
	case context.DeadlineExceeded:
		return status.Error(codes.DeadlineExceeded, "deadline exceeded")
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

// Your usecase returns fresh errors.New("..."). Mirror the same sentinels here.
var (
	contextCanceledErr = errors.New("context cancelled")
	contextTimedOutErr = errors.New("timed out")
)

// Map usecase-level errors to gRPC status codes.
func toUsecaseStatus(err error) error {
	switch {
	case errors.Is(err, contextCanceledErr):
		return status.Error(codes.Canceled, "client canceled request")
	case errors.Is(err, contextTimedOutErr):
		return status.Error(codes.DeadlineExceeded, "deadline exceeded")
	// Add more domain error mappings here, for example:
	// case errors.Is(err, usecase.ErrValidation):
	//     return status.Error(codes.InvalidArgument, err.Error())
	// case errors.Is(err, usecase.ErrNotFound):
	//     return status.Error(codes.NotFound, "not found")
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
