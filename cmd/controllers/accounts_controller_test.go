package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/joolshouston/pismo-technical-test/cmd/services"
	"github.com/joolshouston/pismo-technical-test/shared/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockMongoRepo struct{}

func (m *MockMongoRepo) FindAllTransactionsForAccountID(ctx context.Context, accountID string) ([]model.Transaction, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockMongoRepo) UpdateTransactionByID(ctx context.Context, transactionID string, transaction model.Transaction) error {
	//TODO implement me
	panic("implement me")
}

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

func Test_CreatAccount(t *testing.T) {
	repo := &MockMongoRepo{}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	accountService := services.NewAccountsService(repo, logger)
	accountsController := NewAccountsController(accountService, logger)
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response, expectedStatus int)
	}{
		{
			name:           "Valid account",
			requestBody:    `{"document_number":"123456789"}`,
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int) {
				if resp.StatusCode != expectedStatus {
					t.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
				var accountResp model.AccountResponseBody
				if err := json.NewDecoder(resp.Body).Decode(&accountResp); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				if accountResp.DocumentNumber != "123456789" {
					t.Errorf("expected document number '123456789', got %s", accountResp.DocumentNumber)
				}
			},
		},
		{
			name:           "Invalid account - empty document number",
			requestBody:    `{"document_number":""}`,
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int) {
				if resp.StatusCode != expectedStatus {
					t.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
				var errResp model.ErrorResponse
				if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				if errResp.Message != "document_number is required" {
					t.Errorf("expected error message 'document_number is required', got %s", errResp.Message)
				}
			},
		},
		{
			name:           "Invalid request body",
			requestBody:    `{document_number":"invalid_json"}`,
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int) {
				if resp.StatusCode != expectedStatus {
					t.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
				var errResp model.ErrorResponse
				if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				if errResp.Message != "invalid request body" {
					t.Errorf("expected error message 'invalid request body', got %s", errResp.Message)
				}
			},
		},
		{
			name:           "Failed to create account",
			requestBody:    `{"document_number":"1234567895"}`,
			expectedStatus: http.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int) {
				if resp.StatusCode != expectedStatus {
					t.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
				var errResp model.ErrorResponse
				if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				if errResp.Message != "failed to create account" {
					t.Errorf("expected error message 'failed to create account', got %s", errResp.Message)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			accountsController.CreateAccount(w, req)
			resp := w.Result()
			tt.validate(t, resp, tt.expectedStatus)
		})
	}
}

func Test_GetAccount(t *testing.T) {
	repo := &MockMongoRepo{}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	accountService := services.NewAccountsService(repo, logger)
	accountsController := NewAccountsController(accountService, logger)

	tests := []struct {
		name           string
		accountID      string
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response, expectedStatus int)
	}{
		{
			name:           "Valid account ID",
			accountID:      "valid_id",
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int) {
				if resp.StatusCode != expectedStatus {
					t.Fatalf("expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
				var accountResp model.AccountResponseBody
				if err := json.NewDecoder(resp.Body).Decode(&accountResp); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				if accountResp.DocumentNumber != "123456789" {
					t.Errorf("expected document number '123456789', got %s", accountResp.DocumentNumber)
				}
			},
		},
		{
			name:           "error getting account",
			accountID:      "invalid_id",
			expectedStatus: http.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int) {
				if resp.StatusCode != expectedStatus {
					t.Fatalf("expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
				var errResp model.ErrorResponse
				if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				if errResp.Message != "failed to get account" {
					t.Errorf("expected error message 'failed to get account', got %s", errResp.Message)
				}
			},
		},
		{
			name:           "missing account id in path",
			accountID:      "",
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int) {
				if resp.StatusCode != expectedStatus {
					t.Fatalf("expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
				var errResp model.ErrorResponse
				if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				if errResp.Message != "account ID is required" {
					t.Errorf("expected error message 'account ID is required', got %s", errResp.Message)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/accounts/"+tt.accountID, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.accountID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			accountsController.GetAccount(w, req)
			resp := w.Result()
			tt.validate(t, resp, tt.expectedStatus)
		})
	}
}
