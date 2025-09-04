package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/joolshouston/pismo-technical-test/shared/model"
	"github.com/joolshouston/pismo-technical-test/shared/repository"
)

type TransactionsInferface interface {
	CreateTransaction(ctx context.Context, transaction model.TransactionRequestBody, idempotencyKey string) (*model.TransactionResponseBody, error)
}

type TransactionService struct {
	repo   repository.DatabaseRepository
	logger *slog.Logger
}

func NewTransactionService(repo repository.DatabaseRepository, logger *slog.Logger) *TransactionService {
	logger.InfoContext(context.Background(), "TransactionService initialized")
	return &TransactionService{repo: repo, logger: logger}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, transaction model.TransactionRequestBody, idempotencyKey string) (*model.TransactionResponseBody, error) {
	s.logger.InfoContext(ctx, "creating transaction", "accountID", transaction.AccountID, "operationTypeID", transaction.OperationID.String(), "amount", transaction.Amount)

	// Check if a transaction with the same idempotency key exists
	existingTx, err := s.repo.FindTransactionByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to check existing transaction by idempotency key", "error", err)
		return nil, errors.New("failed to check existing transaction by idempotency key")
	}
	if existingTx != nil {
		s.logger.InfoContext(ctx, "transaction with the same idempotency key already exists", "idempotencyKey", idempotencyKey)
		return &model.TransactionResponseBody{
			TransactionID: existingTx.ID.Hex(),
			AccountID:     existingTx.AccountID,
			OperationID:   existingTx.OperationID,
			Amount:        existingTx.Amount,
		}, nil
	}
	// Validate request in service layer
	if transaction.OperationID.String() == "UNKNOWN" {
		return nil, errors.New("invalid operation type")
	}

	// installment purchase, installment purchase and withdrawal should be negative amounts
	if (transaction.OperationID == model.OperationTypePurchase || transaction.OperationID == model.OperationTypeInstallmentPurchase || transaction.OperationID == model.OperationTypeWithdrawal) && transaction.Amount > 0 {
		return nil, errors.New("invalid operation type for payment amount")
	}

	// payment should be positive amount
	if transaction.OperationID == model.OperationTypePayment && transaction.Amount < 0 {
		return nil, errors.New("invalid operation type for payment amount")
	}

	// payment should be positive amount
	if transaction.OperationID == model.OperationTypePayment && transaction.Amount < 0 {
		return nil, errors.New("invalid operation type for payment amount")
	}

	// Check if account exists
	account, err := s.repo.GetAccountByID(ctx, transaction.AccountID)
	if err != nil {
		return nil, errors.New("account query failed")
	}
	if account == nil {
		return nil, errors.New("account not found")
	}
	tx := model.Transaction{
		AccountID:   transaction.AccountID,
		OperationID: transaction.OperationID,
		Amount:      transaction.Amount,
		EventDate:   time.Now().Format(time.RFC3339Nano),
	}

	// I am wondering whether it would make sense to ALWAYS save the transaction with the idempotency key even if the request is invalid
	// For now I will not do this but validate the if a record with the idempotency key exists first and fail before it reaches here
	createdTx, err := s.repo.CreateTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}
	return &model.TransactionResponseBody{
		TransactionID: createdTx.ID.Hex(),
		AccountID:     createdTx.AccountID,
		OperationID:   createdTx.OperationID,
		Amount:        createdTx.Amount,
	}, nil
}
