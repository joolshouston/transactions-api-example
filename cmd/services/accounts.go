package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/joolshouston/pismo-technical-test/shared/model"
	"github.com/joolshouston/pismo-technical-test/shared/repository"
)

type AccountsInterface interface {
	CreateAccount(ctx context.Context, documentID string) (*model.AccountResponseBody, error)
	GetAccountByID(ctx context.Context, accountID string) (*model.AccountResponseBody, error)
	CreateTransaction(ctx context.Context, transaction model.TransactionRequestBody) (*model.TransactionResponseBody, error)
}

type AccountsService struct {
	repo   repository.DatabaseRepository
	logger *slog.Logger
}

func NewAccountsService(repo repository.DatabaseRepository, logger *slog.Logger) *AccountsService {
	logger.InfoContext(context.Background(), "AccountsService initialized")
	return &AccountsService{repo: repo, logger: logger}
}

func (s *AccountsService) CreateAccount(ctx context.Context, documentID string) (*model.AccountResponseBody, error) {
	s.logger.InfoContext(ctx, "creating account", "documentID", documentID)
	existingAccount, err := s.repo.GetAccountByDocumentNumber(ctx, documentID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to check account existence", "error", err)
		return nil, err
	}
	if existingAccount != nil {
		return nil, fmt.Errorf("account already exists")
	}
	acc, err := s.repo.CreateAccount(ctx, documentID)
	if err != nil {
		return nil, err
	}
	return &model.AccountResponseBody{
		AccountID:      acc.ID.Hex(),
		DocumentNumber: acc.DocumentNumber,
	}, nil
}

func (s *AccountsService) GetAccountByID(ctx context.Context, accountID string) (*model.AccountResponseBody, error) {
	acc, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return &model.AccountResponseBody{
		AccountID:      acc.ID.Hex(),
		DocumentNumber: acc.DocumentNumber,
	}, nil
}
