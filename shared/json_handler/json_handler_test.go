package json_handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joolshouston/pismo-technical-test/shared/model"
)

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		data           interface{}
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response)
	}{
		{
			name:           "Write successful JSON response",
			status:         http.StatusOK,
			data:           map[string]string{"message": "success"},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *http.Response) {
				if resp.Header.Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type 'application/json', got %s", resp.Header.Get("Content-Type"))
				}
				var result map[string]string
				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatalf("expected no error decoding response, got %v", err)
				}
				if result["message"] != "success" {
					t.Errorf("expected message 'success', got %s", result["message"])
				}
			},
		},
		{
			name:           "Write JSON response with custom status",
			status:         http.StatusCreated,
			data:           model.AccountResponseBody{AccountID: "12345", DocumentNumber: "123456789"},
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, resp *http.Response) {
				if resp.Header.Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type 'application/json', got %s", resp.Header.Get("Content-Type"))
				}
				var result model.AccountResponseBody
				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatalf("expected no error decoding response, got %v", err)
				}
				if result.AccountID != "12345" {
					t.Errorf("expected account ID '12345', got %s", result.AccountID)
				}
				if result.DocumentNumber != "123456789" {
					t.Errorf("expected document number '123456789', got %s", result.DocumentNumber)
				}
			},
		},
		{
			name:           "Write JSON response with nil data",
			status:         http.StatusNoContent,
			data:           nil,
			expectedStatus: http.StatusNoContent,
			validate: func(t *testing.T, resp *http.Response) {
				if resp.Header.Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type 'application/json', got %s", resp.Header.Get("Content-Type"))
				}
			},
		},
		{
			name:           "Write JSON response with complex data",
			status:         http.StatusOK,
			data:           model.TransactionResponseBody{TransactionID: "txn123", AccountID: "acc456", OperationID: 1, Amount: -12350},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *http.Response) {
				if resp.Header.Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type 'application/json', got %s", resp.Header.Get("Content-Type"))
				}
				var result model.TransactionResponseBody
				err := json.NewDecoder(resp.Body).Decode(&result)
				if err != nil {
					t.Fatalf("expected no error decoding response, got %v", err)
				}
				if result.TransactionID != "txn123" {
					t.Errorf("expected transaction ID 'txn123', got %s", result.TransactionID)
				}
				if result.Amount != -12350 {
					t.Errorf("expected amount -12350, got %f", result.Amount)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			WriteJSON(resp, tt.status, tt.data)

			result := resp.Result()
			if result.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, result.StatusCode)
			}
			tt.validate(t, result)
		})
	}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		errMsg         string
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response)
	}{
		{
			name:           "Write bad request error",
			status:         http.StatusBadRequest,
			errMsg:         "invalid request body",
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, resp *http.Response) {
				if resp.Header.Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type 'application/json', got %s", resp.Header.Get("Content-Type"))
				}
				var errResp model.ErrorResponse
				err := json.NewDecoder(resp.Body).Decode(&errResp)
				if err != nil {
					t.Fatalf("expected no error decoding response, got %v", err)
				}
				if errResp.Message != "invalid request body" {
					t.Errorf("expected error message 'invalid request body', got %s", errResp.Message)
				}
				if errResp.Status != http.StatusBadRequest {
					t.Errorf("expected status %d, got %d", http.StatusBadRequest, errResp.Status)
				}
			},
		},
		{
			name:           "Write internal server error",
			status:         http.StatusInternalServerError,
			errMsg:         "database connection failed",
			expectedStatus: http.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response) {
				if resp.Header.Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type 'application/json', got %s", resp.Header.Get("Content-Type"))
				}
				var errResp model.ErrorResponse
				err := json.NewDecoder(resp.Body).Decode(&errResp)
				if err != nil {
					t.Fatalf("expected no error decoding response, got %v", err)
				}
				if errResp.Message != "database connection failed" {
					t.Errorf("expected error message 'database connection failed', got %s", errResp.Message)
				}
				if errResp.Status != http.StatusInternalServerError {
					t.Errorf("expected status %d, got %d", http.StatusInternalServerError, errResp.Status)
				}
			},
		},
		{
			name:           "Write not found error",
			status:         http.StatusNotFound,
			errMsg:         "resource not found",
			expectedStatus: http.StatusNotFound,
			validate: func(t *testing.T, resp *http.Response) {
				if resp.Header.Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type 'application/json', got %s", resp.Header.Get("Content-Type"))
				}
				var errResp model.ErrorResponse
				err := json.NewDecoder(resp.Body).Decode(&errResp)
				if err != nil {
					t.Fatalf("expected no error decoding response, got %v", err)
				}
				if errResp.Message != "resource not found" {
					t.Errorf("expected error message 'resource not found', got %s", errResp.Message)
				}
				if errResp.Status != http.StatusNotFound {
					t.Errorf("expected status %d, got %d", http.StatusNotFound, errResp.Status)
				}
			},
		},
		{
			name:           "Write unauthorized error",
			status:         http.StatusUnauthorized,
			errMsg:         "authentication required",
			expectedStatus: http.StatusUnauthorized,
			validate: func(t *testing.T, resp *http.Response) {
				if resp.Header.Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type 'application/json', got %s", resp.Header.Get("Content-Type"))
				}
				var errResp model.ErrorResponse
				err := json.NewDecoder(resp.Body).Decode(&errResp)
				if err != nil {
					t.Fatalf("expected no error decoding response, got %v", err)
				}
				if errResp.Message != "authentication required" {
					t.Errorf("expected error message 'authentication required', got %s", errResp.Message)
				}
				if errResp.Status != http.StatusUnauthorized {
					t.Errorf("expected status %d, got %d", http.StatusUnauthorized, errResp.Status)
				}
			},
		},
		{
			name:           "Write error with empty message",
			status:         http.StatusBadRequest,
			errMsg:         "",
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, resp *http.Response) {
				if resp.Header.Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type 'application/json', got %s", resp.Header.Get("Content-Type"))
				}
				var errResp model.ErrorResponse
				err := json.NewDecoder(resp.Body).Decode(&errResp)
				if err != nil {
					t.Fatalf("expected no error decoding response, got %v", err)
				}
				if errResp.Message != "" {
					t.Errorf("expected empty error message, got %s", errResp.Message)
				}
				if errResp.Status != http.StatusBadRequest {
					t.Errorf("expected status %d, got %d", http.StatusBadRequest, errResp.Status)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := httptest.NewRecorder()
			errorResponse := &model.ErrorResponse{
				Message: tt.errMsg,
				Status:  tt.status,
			}
			WriteError(resp, errorResponse)

			result := resp.Result()
			if result.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, result.StatusCode)
			}
			tt.validate(t, result)
		})
	}
}
