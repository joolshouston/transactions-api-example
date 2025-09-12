package repository

import (
	"context"

	"github.com/joolshouston/pismo-technical-test/shared/model"
)

type DatabaseRepository interface {
	CreateAccount(ctx context.Context, documentID string) (*model.Account, error)
	GetAccountByID(ctx context.Context, accountID string) (*model.Account, error)
	GetAccountByDocumentNumber(ctx context.Context, documentNumber string) (*model.Account, error)
	CreateTransaction(ctx context.Context, transaction model.Transaction) (*model.Transaction, error)
	FindTransactionByIdempotencyKey(ctx context.Context, idempotencyKey string) (*model.Transaction, error)
	FindAllTransactionsForAccountID(ctx context.Context, accountID string) ([]model.Transaction, error)
	UpdateTransactionByID(ctx context.Context, transactionID string, transaction model.Transaction) error
}
