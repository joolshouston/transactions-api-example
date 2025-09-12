package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/joolshouston/pismo-technical-test/cmd/controllers"
	"github.com/joolshouston/pismo-technical-test/cmd/services"
	"github.com/joolshouston/pismo-technical-test/shared/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockRouteRepo struct{}

func (m *MockRouteRepo) GetAccountByID(ctx context.Context, accountID string) (*model.Account, error) {
	switch accountID {
	case "valid_id":
		return &model.Account{
			ID:             bson.NewObjectID(),
			DocumentNumber: "123456789",
		}, nil
	case "invalid_id":
		return nil, nil
	default:
		return nil, nil
	}
}

func (m *MockRouteRepo) GetAccountByDocumentNumber(ctx context.Context, documentNumber string) (*model.Account, error) {
	switch documentNumber {
	case "valid_document_number":
		return &model.Account{
			ID:             bson.NewObjectID(),
			DocumentNumber: documentNumber,
		}, nil
	case "invalid_document_number":
		return nil, errors.New("invalid document number")
	default:
		return nil, nil
	}
}

func (m *MockRouteRepo) CreateAccount(ctx context.Context, documentID string) (*model.Account, error) {
	return &model.Account{
		ID:             bson.NewObjectID(),
		DocumentNumber: documentID,
	}, nil
}

func (m *MockRouteRepo) FindAccountByID(ctx context.Context, accountID string) (*model.Account, error) {
	switch accountID {
	case "valid_id":
		return &model.Account{
			ID:             bson.NewObjectID(),
			DocumentNumber: "123456789",
		}, nil
	case "invalid_id":
		return nil, errors.New("invalid account id")
	default:
		return nil, nil
	}
}
func (m *MockRouteRepo) CreateTransaction(ctx context.Context, transaction model.Transaction) (*model.Transaction, error) {
	return &model.Transaction{
		ID:          bson.NewObjectID(),
		AccountID:   transaction.AccountID,
		OperationID: transaction.OperationID,
		Amount:      transaction.Amount,
	}, nil
}

func (m *MockRouteRepo) FindTransactionByIdempotencyKey(ctx context.Context, idempotencyKey string) (*model.Transaction, error) {
	return nil, nil
}

func (m *MockRouteRepo) FindAllTransactionsForAccountID(ctx context.Context, accountID string) ([]model.Transaction, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockRouteRepo) UpdateTransactionByID(ctx context.Context, transactionID string, transaction model.Transaction) error {
	//TODO implement me
	panic("implement me")
}

func TestRoutes(t *testing.T) {
	repo := &MockRouteRepo{}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	accountService := services.NewAccountsService(repo, logger)
	transactionService := services.NewTransactionService(repo, logger)
	accountController := controllers.NewAccountsController(accountService, logger)
	transactionController := controllers.NewTransactionsController(transactionService, logger)

	app := &Application{}
	router := app.routes(accountController, transactionController)

	tests := []struct {
		name           string
		method         string
		url            string
		body           string
		headers        map[string]string
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response, expectedStatus int)
	}{
		{
			name:           "POST /v1/accounts - successful account creation",
			method:         "POST",
			url:            "/v1/accounts",
			body:           `{"document_number":"123456789"}`,
			headers:        map[string]string{"Content-Type": "application/json"},
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int) {
				var account model.AccountResponseBody
				err := json.NewDecoder(resp.Body).Decode(&account)
				if err != nil {
					t.Fatalf("expected no error decoding response, got %v", err)
				}
				if account.DocumentNumber != "123456789" {
					t.Errorf("expected document number '123456789', got %s", account.DocumentNumber)
				}
			},
		},
		{
			name:           "POST /v1/accounts - invalid request body",
			method:         "POST",
			url:            "/v1/accounts",
			body:           `{"invalid": "data"}`,
			headers:        map[string]string{"Content-Type": "application/json"},
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int) {
				var errResp model.ErrorResponse
				err := json.NewDecoder(resp.Body).Decode(&errResp)
				if err != nil {
					t.Fatalf("expected no error decoding response, got %v", err)
				}
				if !strings.Contains(errResp.Message, "document_number") {
					t.Errorf("expected error message to contain 'document_number', got %s", errResp.Message)
				}
			},
		},
		{
			name:           "GET /v1/accounts/{id} - successful account retrieval",
			method:         "GET",
			url:            "/v1/accounts/valid_id",
			body:           "",
			headers:        map[string]string{},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int) {
				var account model.AccountResponseBody
				err := json.NewDecoder(resp.Body).Decode(&account)
				if err != nil {
					t.Fatalf("expected no error decoding response, got %v", err)
				}
				if account.DocumentNumber != "123456789" {
					t.Errorf("expected document number '123456789', got %s", account.DocumentNumber)
				}
			},
		},
		{
			name:           "GET /v1/accounts/{id} - account not found",
			method:         "GET",
			url:            "/v1/accounts/invalid_id",
			body:           "",
			headers:        map[string]string{},
			expectedStatus: http.StatusNotFound,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int) {
				var errResp model.ErrorResponse
				err := json.NewDecoder(resp.Body).Decode(&errResp)
				if err != nil {
					t.Fatalf("expected no error decoding response, got %v", err)
				}
				if !strings.Contains(errResp.Message, "not found") {
					t.Errorf("expected error message to contain 'not found', got %s", errResp.Message)
				}
			},
		},
		{
			name:   "POST /v1/transactions - successful transaction creation",
			method: "POST",
			url:    "/v1/transactions",
			body:   `{"account_id":"valid_id","operation_type_id":1,"amount":-123.5}`,
			headers: map[string]string{
				"Content-Type":      "application/json",
				"X-idempotency-Key": "unique-key-123",
			},
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int) {
				var transaction model.TransactionResponseBody
				err := json.NewDecoder(resp.Body).Decode(&transaction)
				if err != nil {
					t.Fatalf("expected no error decoding response, got %v", err)
				}
				if transaction.AccountID != "valid_id" {
					t.Errorf("expected account ID 'valid_id', got %s", transaction.AccountID)
				}
				if transaction.Amount != -123.5 {
					t.Errorf("expected amount -123.5, got %f", transaction.Amount)
				}
			},
		},
		{
			name:           "POST /v1/transactions - missing idempotency key",
			method:         "POST",
			url:            "/v1/transactions",
			body:           `{"account_id":"valid_id","operation_type_id":1,"amount":-123.5}`,
			headers:        map[string]string{"Content-Type": "application/json"},
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int) {
				if resp.StatusCode != expectedStatus {
					t.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
				var errResp model.ErrorResponse
				err := json.NewDecoder(resp.Body).Decode(&errResp)
				if err != nil {
					t.Fatalf("expected no error decoding response, got %v", err)
				}
				if !strings.Contains(errResp.Message, "idempotency") {
					t.Errorf("expected error message to contain 'idempotency', got %s", errResp.Message)
				}
			},
		},
		{
			name:           "POST /v1/transactions - invalid request body",
			method:         "POST",
			url:            "/v1/transactions",
			body:           `{"invalid": "data"}`,
			headers:        map[string]string{"Content-Type": "application/json", "X-idempotency-Key": "test-key"},
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int) {
				if resp.StatusCode != expectedStatus {
					t.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
				var errResp model.ErrorResponse
				err := json.NewDecoder(resp.Body).Decode(&errResp)
				if err != nil {
					t.Fatalf("expected no error decoding response, got %v", err)
				}
				if !strings.Contains(errResp.Message, "missing") || !strings.Contains(errResp.Message, "account_id") {
					t.Errorf("expected error message to contain missing fields, got %s", errResp.Message)
				}
			},
		},
		{
			name:           "POST /invalid-route - route not found",
			method:         "POST",
			url:            "/invalid-route",
			body:           `{}`,
			headers:        map[string]string{"Content-Type": "application/json"},
			expectedStatus: http.StatusNotFound,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int) {
				if resp.StatusCode != expectedStatus {
					t.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
			},
		},
		{
			name:           "GET /v1/invalid-endpoint - endpoint not found",
			method:         "GET",
			url:            "/v1/invalid-endpoint",
			body:           "",
			headers:        map[string]string{},
			expectedStatus: http.StatusNotFound,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int) {
				if resp.StatusCode != expectedStatus {
					t.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.url, bytes.NewBufferString(tt.body))
			} else {
				req = httptest.NewRequest(tt.method, tt.url, nil)
			}
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			result := resp.Result()

			tt.validate(t, result, tt.expectedStatus)
		})
	}
}

func TestRoutesMiddleware(t *testing.T) {
	repo := &MockRouteRepo{}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	accountService := services.NewAccountsService(repo, logger)
	transactionService := services.NewTransactionService(repo, logger)
	accountController := controllers.NewAccountsController(accountService, logger)
	transactionController := controllers.NewTransactionsController(transactionService, logger)

	app := &Application{}
	router := app.routes(accountController, transactionController)

	tests := []struct {
		name           string
		method         string
		url            string
		body           string
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response, expectedStatus int)
	}{
		{
			name:           "Test timeout middleware is applied",
			method:         "POST",
			url:            "/v1/accounts",
			body:           `{"document_number":"123456789"}`,
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int) {
				if resp.StatusCode != expectedStatus {
					t.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
				var account model.AccountResponseBody
				err := json.NewDecoder(resp.Body).Decode(&account)
				if err != nil {
					t.Fatalf("expected no error decoding response, got %v", err)
				}
			},
		},
		{
			name:           "Test recoverer middleware handles panics",
			method:         "GET",
			url:            "/v1/accounts/trigger-panic",
			body:           "",
			expectedStatus: http.StatusNotFound,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int) {
				if resp.StatusCode != expectedStatus {
					t.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.url, bytes.NewBufferString(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.url, nil)
			}

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			result := resp.Result()

			tt.validate(t, result, tt.expectedStatus)
		})
	}
}
