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

	"github.com/joolshouston/pismo-technical-test/cmd/services"
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
			Amount:      transaction.Amount,
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

func Test_CreateTransaction(t *testing.T) {
	repo := &MockMongoRepo{}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	service := services.NewTransactionService(repo, logger)
	transactionController := NewTransactionsController(service, logger)

	tests := []struct {
		name           string
		transaction    string
		idempotencyKey string
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response, expectedStatus int)
	}{
		{
			name:           "Successful transaction creation",
			transaction:    `{"account_id":"valid_id","operation_type_id":1,"amount":-123.5}`,
			idempotencyKey: "x-idempotency-key-unique",
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int) {
				if resp.StatusCode != expectedStatus {
					t.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
				var respBody model.TransactionResponseBody
				err := json.NewDecoder(resp.Body).Decode(&respBody)

				// resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset body for further reading if needed
				if err != nil {
					t.Fatalf("expected no error decoding response, got %v", err)
				}
				if respBody.AccountID != "valid_id" {
					t.Errorf("expected account ID 'valid_id', got %s", respBody.AccountID)
				}
				if respBody.OperationID != 1 {
					t.Errorf("expected operation ID 1, got %d", respBody.OperationID)
				}
				if respBody.Amount != -123.5 {
					t.Errorf("expected amount -123.5, got %f", respBody.Amount)
				}
			},
		},
		{
			name:           "Duplicate idempotency key used again, should return existing transaction",
			transaction:    `{"account_id":"valid_id","operation_type_id":1,"amount":-123.5}`,
			idempotencyKey: "x-idempotency-key-duplicate",
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int) {
				if resp.StatusCode != expectedStatus {
					t.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
				var respBody model.TransactionResponseBody
				err := json.NewDecoder(resp.Body).Decode(&respBody)
				if err != nil {
					t.Fatalf("expected no error decoding response, got %v", err)
				}
				if respBody.AccountID != "valid_id" {
					t.Errorf("expected account ID 'valid_id', got %s", respBody.AccountID)
				}
				if respBody.OperationID != 1 {
					t.Errorf("expected operation ID 1, got %d", respBody.OperationID)
				}
				if respBody.Amount != -123.5 {
					t.Errorf("expected amount -123.5, got %f", respBody.Amount)
				}
			},
		},
		{
			name:           "Invalid operation type",
			transaction:    `{"account_id":"valid_id","operation_type_id":99,"amount":-123.5}`,
			idempotencyKey: "x-idempotency-key-unique",
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
				if errResp.Message != "invalid operation type" {
					t.Errorf("expected error message 'failed to create transaction: invalid operation type', got %s", errResp.Message)
				}
			},
		},
		{
			name:           "Idempotency key missing",
			transaction:    `{"account_id":"valid_id","operation_type_id":1,"amount":-123.5}`,
			idempotencyKey: "",
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
				if errResp.Message != "X-idempotency-Key header is required" {
					t.Errorf("expected error message 'X-idempotency-Key header is required', got %s", errResp.Message)
				}
			},
		},
		{
			name:           "Invalid request body",
			transaction:    `{"account_id":"valn_type_id":1,"amount":-123.5}`,
			idempotencyKey: "x-idempotency-key-unique",
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
				if errResp.Message != "invalid request body" {
					t.Errorf("expected error message 'invalid request body', got %s", errResp.Message)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(tt.transaction))
			req.Header.Set("Content-Type", "application/json")
			if tt.idempotencyKey != "" {
				req.Header.Set("x-idempotency-key", tt.idempotencyKey)
			}
			transactionController.CreateTransaction(resp, req)
			tt.validate(t, resp.Result(), tt.expectedStatus)
		})
	}
}
