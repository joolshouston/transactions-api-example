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

type MockMongoRepo struct{}

func (m *MockMongoRepo) CreateAccount(ctx context.Context, documentID string) (*model.Account, error) {
	switch documentID {
	case "":
		return nil, errors.New("invalid document ID")
	case "1234567895":
		return nil, errors.New("duplicate document ID")
	default:
		return &model.Account{
			ID:             bson.NewObjectID(),
			DocumentNumber: documentID,
		}, nil
	}
}

func (m *MockMongoRepo) GetAccountByID(ctx context.Context, accountID string) (*model.Account, error) {
	switch accountID {
	case "valid_id":
		return &model.Account{
			ID:             bson.NewObjectID(),
			DocumentNumber: "123456789",
		}, nil
	case "invalid_id":
		return nil, errors.New("account not found")
	case "account_nonexistent":
		return nil, nil
	default:
		return &model.Account{
			ID:             bson.NewObjectID(),
			DocumentNumber: "1",
		}, nil
	}
}

func (m *MockMongoRepo) GetAccountByDocumentNumber(ctx context.Context, documentNumber string) (*model.Account, error) {
	switch documentNumber {
	case "failed_to_get_account":
		return &model.Account{
			ID:             bson.NewObjectID(),
			DocumentNumber: documentNumber,
		}, nil
	case "":
		return nil, errors.New("invalid document number")
	default:
	}
	return nil, nil
}

func Test_CreateAccount(t *testing.T) {
	repo := &MockMongoRepo{}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	accountService := NewAccountsService(repo, logger)
	tests := []struct {
		name        string
		requestBody model.AccountRequestBody
		validate    func(t *testing.T, resp *model.AccountResponseBody, err *model.ErrorResponse)
	}{
		{
			name: "Valid account",
			requestBody: model.AccountRequestBody{
				DocumentNumber: "123456789",
			},
			validate: func(t *testing.T, resp *model.AccountResponseBody, err *model.ErrorResponse) {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if resp.DocumentNumber != "123456789" {
					t.Errorf("expected document number '123456789', got %s", resp.DocumentNumber)
				}
			},
		},
		{
			name: "Invalid account",
			requestBody: model.AccountRequestBody{
				DocumentNumber: "",
			},
			validate: func(t *testing.T, resp *model.AccountResponseBody, err *model.ErrorResponse) {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			},
		},
		{
			name: "Duplicate account",
			requestBody: model.AccountRequestBody{
				DocumentNumber: "1234567895",
			},
			validate: func(t *testing.T, resp *model.AccountResponseBody, err *model.ErrorResponse) {
				if err == nil {
					t.Fatalf("expected error for duplicate account, got nil")
				}
			},
		},
		{
			name: "Failed to get account",
			requestBody: model.AccountRequestBody{
				DocumentNumber: "failed_to_get_account",
			},
			validate: func(t *testing.T, resp *model.AccountResponseBody, err *model.ErrorResponse) {
				if err == nil {
					t.Fatalf("expected error for failed to get account, got nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := accountService.CreateAccount(context.Background(), tt.requestBody.DocumentNumber)
			tt.validate(t, resp, err)
		})
	}
}

func Test_GetAccountByID(t *testing.T) {
	repo := &MockMongoRepo{}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	accountService := NewAccountsService(repo, logger)
	tests := []struct {
		name      string
		accountID string
		validate  func(t *testing.T, resp *model.AccountResponseBody, err *model.ErrorResponse)
	}{
		{
			name:      "Valid account ID",
			accountID: "valid_id",
			validate: func(t *testing.T, resp *model.AccountResponseBody, err *model.ErrorResponse) {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if resp.DocumentNumber != "123456789" {
					t.Errorf("expected document number '123456789', got %s", resp.DocumentNumber)
				}
			},
		},
		{
			name:      "Invalid account ID",
			accountID: "invalid_id",
			validate: func(t *testing.T, resp *model.AccountResponseBody, err *model.ErrorResponse) {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := accountService.GetAccountByID(context.Background(), tt.accountID)
			tt.validate(t, resp, err)
		})
	}
}
