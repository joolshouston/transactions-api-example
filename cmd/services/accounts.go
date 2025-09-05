package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/joolshouston/pismo-technical-test/shared/model"
	"github.com/joolshouston/pismo-technical-test/shared/repository"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AccountsInterface interface {
	CreateAccount(ctx context.Context, documentID string) (*model.AccountResponseBody, *model.ErrorResponse)
	GetAccountByID(ctx context.Context, accountID string) (*model.AccountResponseBody, *model.ErrorResponse)
}

type AccountsService struct {
	repo   repository.DatabaseRepository
	logger *slog.Logger
}

func NewAccountsService(repo repository.DatabaseRepository, logger *slog.Logger) *AccountsService {
	logger.InfoContext(context.Background(), "AccountsService initialized")
	return &AccountsService{repo: repo, logger: logger}
}

func (s *AccountsService) CreateAccount(ctx context.Context, documentID string) (*model.AccountResponseBody, *model.ErrorResponse) {
	s.logger.InfoContext(ctx, "creating account", "documentID", documentID)
	existingAccount, err := s.repo.GetAccountByDocumentNumber(ctx, documentID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to check account existence", "error", err)
		return nil, &model.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: fmt.Sprintf("failed to check account existence"),
		}
	}
	if existingAccount != nil {
		s.logger.InfoContext(ctx, "account already exists")
		return nil, &model.ErrorResponse{
			Status:  http.StatusConflict,
			Message: fmt.Sprintf("account already exists"),
		}
	}
	acc, err := s.repo.CreateAccount(ctx, documentID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create account", "error", err)
		return nil, &model.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: fmt.Sprintf("failed to create account"),
		}
	}
	return &model.AccountResponseBody{
		AccountID:      acc.ID.Hex(),
		DocumentNumber: acc.DocumentNumber,
	}, nil
}

func (s *AccountsService) GetAccountByID(ctx context.Context, accountID string) (*model.AccountResponseBody, *model.ErrorResponse) {
	acc, err := s.repo.GetAccountByID(ctx, accountID)
	if errors.Is(err, mongo.ErrNoDocuments) {
		s.logger.ErrorContext(ctx, "failed to get account", "error", err)
		return nil, &model.ErrorResponse{
			Status:  http.StatusNotFound,
			Message: fmt.Sprintf("account not found"),
		}
	}
	if err != nil {
		return nil, &model.ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: fmt.Sprintf("failed to get account"),
		}
	}
	if acc == nil {
		s.logger.InfoContext(ctx, "account not found")
		return nil, &model.ErrorResponse{
			Status:  http.StatusNotFound,
			Message: "account not found",
		}
	}
	return &model.AccountResponseBody{
		AccountID:      acc.ID.Hex(),
		DocumentNumber: acc.DocumentNumber,
	}, nil
}
