package services

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/joolshouston/pismo-technical-test/shared/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (m *MockMongoRepo) CreateTransaction(ctx context.Context, transaction model.Transaction) (*model.Transaction, error) {
	switch transaction.Amount {
	case -123.5:
		return &model.Transaction{
			ID:          bson.NewObjectID(),
			AccountID:   transaction.AccountID,
			OperationID: transaction.OperationID,
		}, nil
	case -99999.0:
		return nil, errors.New("transaction creation failed")
	default:
		return &model.Transaction{
			ID:          bson.NewObjectID(),
			AccountID:   transaction.AccountID,
			OperationID: transaction.OperationID,
		}, nil
	}
}

func (m *MockMongoRepo) FindTransactionByIdempotencyKey(ctx context.Context, idempotencyKey string) (*model.Transaction, error) {
	switch idempotencyKey {
	case "x-idempotency-key-duplicate":
		return &model.Transaction{
			ID:          bson.NewObjectID(),
			AccountID:   "valid_id",
			OperationID: 1,
			Amount:      -123.5,
		}, nil
	case "x-idempotency-key-fail":
		return nil, errors.New("database error")
	default:
		return nil, nil
	}
}

func (m *MockMongoRepo) FindAllTransactionsForAccountID(ctx context.Context, accountID string) ([]model.Transaction, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockMongoRepo) UpdateTransactionByID(ctx context.Context, transactionID string, transaction model.Transaction) error {
	//TODO implement me
	panic("implement me")
}

func Test_CreateTransaction(t *testing.T) {
	repo := &MockMongoRepo{}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	service := NewTransactionService(repo, logger)

	tests := []struct {
		name           string
		transaction    model.TransactionRequestBody
		idempotencyKey string
		validate       func(t *testing.T, resp *model.TransactionResponseBody, err *model.ErrorResponse)
	}{
		{
			name: "Successful transaction creation",
			transaction: model.TransactionRequestBody{
				AccountID:   "valid_id",
				OperationID: 1,
				Amount:      -100.0,
			},
			idempotencyKey: "x-idempotency-key-1",
			validate: func(t *testing.T, resp *model.TransactionResponseBody, err *model.ErrorResponse) {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if resp == nil || resp.TransactionID == "" {
					t.Fatalf("expected valid response, got %v", resp)
				}
			},
		},
		{
			name: "Duplicate idempotency key",
			transaction: model.TransactionRequestBody{
				AccountID:   "valid_id",
				OperationID: 1,
				Amount:      -123.5,
			},
			idempotencyKey: "x-idempotency-key-duplicate",
			validate: func(t *testing.T, resp *model.TransactionResponseBody, err *model.ErrorResponse) {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if resp == nil || resp.TransactionID == "" {
					t.Fatalf("expected valid response, got %v", resp)
				}
				if resp.Amount != -123.5 {
					t.Fatalf("expected amount -123.5, got %v", resp.Amount)
				}
				if resp.OperationID != 1 {
					t.Fatalf("expected operation ID 1, got %v", resp.OperationID)
				}
				if resp.AccountID != "valid_id" {
					t.Fatalf("expected account ID 'valid_id', got %v", resp.AccountID)
				}
			},
		},
		{
			name: "Idempotency query fails",
			transaction: model.TransactionRequestBody{
				AccountID:   "valid_id",
				OperationID: 1,
				Amount:      -100.0,
			},
			idempotencyKey: "x-idempotency-key-fail",
			validate: func(t *testing.T, resp *model.TransactionResponseBody, err *model.ErrorResponse) {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			},
		},
		{
			name: "Invalid operation type",
			transaction: model.TransactionRequestBody{
				AccountID:   "valid_id",
				OperationID: 6,
				Amount:      100.0,
			},
			idempotencyKey: "x-idempotency-key-1",
			validate: func(t *testing.T, resp *model.TransactionResponseBody, err *model.ErrorResponse) {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			},
		},
		{
			name: "Account not found",
			transaction: model.TransactionRequestBody{
				AccountID:   "account_nonexistent",
				OperationID: 1,
				Amount:      -100.0,
			},
			idempotencyKey: "x-idempotency-key-1",
			validate: func(t *testing.T, resp *model.TransactionResponseBody, err *model.ErrorResponse) {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if err.Message != "account not found" {
					t.Fatalf("expected error 'account not found', got %v", err.Message)
				}
			},
		},
		{
			name: "Account query fails",
			transaction: model.TransactionRequestBody{
				AccountID:   "invalid_id",
				OperationID: 1,
				Amount:      -100.0,
			},
			idempotencyKey: "x-idempotency-key-1",
			validate: func(t *testing.T, resp *model.TransactionResponseBody, err *model.ErrorResponse) {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if err.Message != "failed to get account" {
					t.Fatalf("expected error 'failed to get account', got %v", err.Message)
				}
			},
		},
		{
			name: "Purchase should be negative amount",
			transaction: model.TransactionRequestBody{
				AccountID:   "valid_id",
				OperationID: 1,
				Amount:      100.0,
			},
			idempotencyKey: "x-idempotency-key-1",
			validate: func(t *testing.T, resp *model.TransactionResponseBody, err *model.ErrorResponse) {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if err.Message != "invalid operation type for transaction amount" {
					t.Fatalf("expected error 'invalid operation type for transaction amount', got %v", err.Message)
				}
			},
		},
		{
			name: "Installment Purchase should be negative amount",
			transaction: model.TransactionRequestBody{
				AccountID:   "valid_id",
				OperationID: 2,
				Amount:      100.0,
			},
			idempotencyKey: "x-idempotency-key-1",
			validate: func(t *testing.T, resp *model.TransactionResponseBody, err *model.ErrorResponse) {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if err.Message != "invalid operation type for transaction amount" {
					t.Fatalf("expected error 'invalid operation type for transaction amount', got %v", err.Message)
				}
			},
		},
		{
			name: "Withdrawal should be negative amount",
			transaction: model.TransactionRequestBody{
				AccountID:   "valid_id",
				OperationID: 3,
				Amount:      100.0,
			},
			idempotencyKey: "x-idempotency-key-1",
			validate: func(t *testing.T, resp *model.TransactionResponseBody, err *model.ErrorResponse) {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if err.Message != "invalid operation type for transaction amount" {
					t.Fatalf("expected error 'invalid operation type for transaction amount', got %v", err.Message)
				}
			},
		},
		{
			name: "Payment should be positive amount",
			transaction: model.TransactionRequestBody{
				AccountID:   "valid_id",
				OperationID: 4,
				Amount:      -100.0,
			},
			idempotencyKey: "x-idempotency-key-1",
			validate: func(t *testing.T, resp *model.TransactionResponseBody, err *model.ErrorResponse) {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if err.Message != "invalid operation type for transaction amount" {
					t.Fatalf("expected error 'invalid operation type for transaction amount', got %v", err.Message)
				}
			},
		},
		{
			name: "Transaction creation fails",
			transaction: model.TransactionRequestBody{
				AccountID:   "valid_id",
				OperationID: 1,
				Amount:      -99999.0,
			},
			idempotencyKey: "x-idempotency-key-1",
			validate: func(t *testing.T, resp *model.TransactionResponseBody, err *model.ErrorResponse) {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if err.Message != "failed to create transaction" {
					t.Fatalf("expected error 'failed to create transaction', got %v", err.Message)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.CreateTransaction(context.Background(), tt.transaction, tt.idempotencyKey)
			tt.validate(t, resp, err)
		})
	}
}
