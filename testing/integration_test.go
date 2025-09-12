//go:build integration
// +build integration

package testing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joolshouston/pismo-technical-test/shared/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	baseURL = "http://localhost:8080/v1"
	timeout = 30 * time.Second
)

var httpClient = &http.Client{
	Timeout: timeout,
}

var mongoClient *mongo.Client

func TestMain(m *testing.M) {
	var err error
	fmt.Println("Running integration tests against live application...")
	clientOptions := options.Client().ApplyURI("mongodb://root:password@localhost:27017/?authSource=admin")
	mongoClient, err = mongo.Connect(clientOptions)
	if err != nil {
		panic(err)
	}

	if err = mongoClient.Ping(context.Background(), nil); err != nil {
		panic(err)
	}

	if !waitForApplication() {
		fmt.Println("Application not available, skipping integration tests...")
		os.Exit(0)
	}

	fmt.Println("Application is ready, running tests...")
	code := m.Run()

	fmt.Println("Integration tests completed")
	os.Exit(code)
}

func waitForApplication() bool {
	maxRetries := 60 // 2 minutes
	for i := 0; i < maxRetries; i++ {
		resp, err := httpClient.Get(baseURL + "/swagger/")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			fmt.Printf("Application ready after %d attempts\n", i+1)
			return true
		}
		if resp != nil {
			resp.Body.Close()
		}

		fmt.Printf("Waiting for application... attempt %d/%d\n", i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}
	return false
}

func TestIntegration_AccountLifecycle(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        *model.AccountRequestBody
		expectedStatusCode int
		validate           func(t *testing.T, requestBody *model.AccountRequestBody, expectedStatusCode int, resp *http.Response, err error)
	}{
		{
			name:               "Create new account successfully",
			expectedStatusCode: http.StatusCreated,
			requestBody: &model.AccountRequestBody{
				DocumentNumber: bson.NewObjectID().Hex(),
			},
			validate: func(t *testing.T, requestBody *model.AccountRequestBody, expectedStatusCode int, resp *http.Response, err error) {
				var account model.AccountResponseBody
				if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if err != nil {
					t.Fatalf("failed to create account: %v", err)
				}
				if resp.StatusCode != expectedStatusCode {
					t.Fatalf("expected status %d, got %d", expectedStatusCode, resp.StatusCode)
				}
				if account.AccountID == "" || account.DocumentNumber != requestBody.DocumentNumber {
					t.Errorf("expected account ID to be set, got %s", account.AccountID)
				}
				if account.DocumentNumber == "" || account.DocumentNumber != requestBody.DocumentNumber {
					t.Errorf("expected document number to be set, got %s", account.DocumentNumber)
				}
			},
		},
		{
			name:               "Create another unique account",
			expectedStatusCode: http.StatusCreated,
			requestBody: &model.AccountRequestBody{
				DocumentNumber: bson.NewObjectID().Hex(),
			},
			validate: func(t *testing.T, requestBody *model.AccountRequestBody, expectedStatusCode int, resp *http.Response, err error) {
				var account model.AccountResponseBody
				if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if err != nil {
					t.Fatalf("failed to create account: %v", err)
				}
				if resp.StatusCode != expectedStatusCode {
					t.Fatalf("expected status %d, got %d", expectedStatusCode, resp.StatusCode)
				}
				if account.AccountID == "" || account.DocumentNumber != requestBody.DocumentNumber {
					t.Errorf("expected account ID to be set, got %s", account.AccountID)
				}
				if account.DocumentNumber == "" || account.DocumentNumber != requestBody.DocumentNumber {
					t.Errorf("expected document number to be set, got %s", account.DocumentNumber)
				}

			},
		},
		{
			name:               "Duplicate account creation should fail",
			expectedStatusCode: http.StatusConflict,
			requestBody: &model.AccountRequestBody{
				DocumentNumber: bson.NewObjectID().Hex(),
			},
			validate: func(t *testing.T, requestBody *model.AccountRequestBody, expectedStatusCode int, resp *http.Response, err error) {
				if resp.StatusCode != http.StatusConflict {
					t.Errorf("expected duplicate account creation to fail with %d, got %d", expectedStatusCode, resp.StatusCode)
				}
				var errorResp model.ErrorResponse
				if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if errorResp.Message != "account already exists" {
					t.Errorf("expected error message to be 'account already exists', got '%s'", errorResp.Message)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *http.Response
			var err error
			accountJSON, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal account request: %v", err)
			}

			resp, err = httpClient.Post(baseURL+"/accounts", "application/json", bytes.NewBuffer(accountJSON))
			if tt.name == "Duplicate account creation should fail" {
				resp, err = httpClient.Post(baseURL+"/accounts", "application/json", bytes.NewBuffer(accountJSON))
			}
			defer resp.Body.Close()
			tt.validate(t, tt.requestBody, tt.expectedStatusCode, resp, err)

		})
	}
}

func TestIntegration_TransactionLifecycle(t *testing.T) {
	accountReq := model.AccountRequestBody{
		DocumentNumber: bson.NewObjectID().Hex(),
	}

	accountJSON, _ := json.Marshal(accountReq)
	resp, err := httpClient.Post(baseURL+"/accounts", "application/json", bytes.NewBuffer(accountJSON))
	if err != nil {
		t.Fatalf("Failed to create test account: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create test account, status: %d", resp.StatusCode)
	}

	var account model.AccountResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		t.Fatalf("Failed to decode account response: %v", err)
	}

	accountID := account.AccountID

	tests := []struct {
		name           string
		transaction    model.TransactionRequestBody
		idempotencyKey string
		expectedStatus int
		validate       func(t *testing.T, transaction model.TransactionRequestBody, expectedStatus int, resp *http.Response, err error)
	}{
		{
			name: "Create purchase transaction",
			transaction: model.TransactionRequestBody{
				AccountID:   accountID,
				OperationID: model.OperationTypePurchase,
				Amount:      -100.50,
			},
			idempotencyKey: "test-purchase-001",
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, transaction model.TransactionRequestBody, expectedStatus int, resp *http.Response, err error) {
				transactionValidator(t, transaction, expectedStatus, resp)
			},
		},
		{
			name: "Create payment transaction",
			transaction: model.TransactionRequestBody{
				AccountID:   accountID,
				OperationID: model.OperationTypePayment,
				Amount:      50.25,
			},
			idempotencyKey: "test-payment-001",
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, transaction model.TransactionRequestBody, expectedStatus int, resp *http.Response, err error) {
				transactionValidator(t, transaction, expectedStatus, resp)
			},
		},
		{
			name: "Create withdrawal transaction",
			transaction: model.TransactionRequestBody{
				AccountID:   accountID,
				OperationID: model.OperationTypeWithdrawal,
				Amount:      -25.75,
			},
			idempotencyKey: "test-withdrawal-001",
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, transaction model.TransactionRequestBody, expectedStatus int, resp *http.Response, err error) {
				transactionValidator(t, transaction, expectedStatus, resp)
			},
		},
		{
			name: "Duplicate idempotency key should return existing transaction",
			transaction: model.TransactionRequestBody{
				AccountID:   accountID,
				OperationID: model.OperationTypePurchase,
				Amount:      -100.50,
			},
			idempotencyKey: "test-purchase-001",
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, transaction model.TransactionRequestBody, expectedStatus int, resp *http.Response, err error) {
				transactionValidator(t, transaction, expectedStatus, resp)
			},
		},
		{
			name: "Missing idempotency key",
			transaction: model.TransactionRequestBody{
				AccountID:   accountID,
				OperationID: model.OperationTypePurchase,
				Amount:      -50.00,
			},
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, transaction model.TransactionRequestBody, expectedStatus int, resp *http.Response, err error) {
				if resp.StatusCode != expectedStatus {
					t.Fatalf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
				var errorResp model.ErrorResponse
				if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if errorResp.Message != "X-idempotency-Key header is required" {
					t.Errorf("Expected error message to be 'X-idempotency-Key header is required', got '%s'", errorResp.Message)
				}

			},
		},
		{
			name: "Invalid operation type",
			transaction: model.TransactionRequestBody{
				AccountID:   accountID,
				OperationID: model.OperationType(99), // Invalid operation
				Amount:      -50.00,
			},
			idempotencyKey: "test-invalid-op",
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, transaction model.TransactionRequestBody, expectedStatus int, resp *http.Response, err error) {
				if resp.StatusCode != expectedStatus {
					t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
				var errorResp model.ErrorResponse
				if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if errorResp.Message != "invalid operation type" {
					t.Errorf("Expected error message to be 'invalid operation type', got '%s'", errorResp.Message)
				}
			},
		},
		{
			name: "Purchase with positive amount should fail",
			transaction: model.TransactionRequestBody{
				AccountID:   accountID,
				OperationID: model.OperationTypePurchase,
				Amount:      50.00, // Should be negative
			},
			idempotencyKey: "test-invalid-purchase",
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, transaction model.TransactionRequestBody, expectedStatus int, resp *http.Response, err error) {
				if resp.StatusCode != expectedStatus {
					t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
				var errorResp model.ErrorResponse
				if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if errorResp.Message != "invalid operation type for transaction amount" {
					t.Errorf("Expected error message to be 'invalid operation type for transaction amount', got '%s'", errorResp.Message)
				}
			},
		},
		{
			name: "Payment with negative amount should fail",
			transaction: model.TransactionRequestBody{
				AccountID:   accountID,
				OperationID: model.OperationTypePayment,
				Amount:      -50.00, // Should be positive
			},
			idempotencyKey: "test-invalid-payment",
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, transaction model.TransactionRequestBody, expectedStatus int, resp *http.Response, err error) {
				if resp.StatusCode != expectedStatus {
					t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
				var errorResp model.ErrorResponse
				if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if errorResp.Message != "invalid operation type for transaction amount" {
					t.Errorf("Expected error message to be 'invalid operation type for transaction amount', got '%s'", errorResp.Message)
				}
			},
		},
		{
			name: "Non-existent account should fail",
			transaction: model.TransactionRequestBody{
				AccountID:   bson.NewObjectID().Hex(), // Non-existent ObjectID
				OperationID: model.OperationTypePurchase,
				Amount:      -50.00,
			},
			idempotencyKey: "test-nonexistent-account",
			expectedStatus: http.StatusNotFound,
			validate: func(t *testing.T, transaction model.TransactionRequestBody, expectedStatus int, resp *http.Response, err error) {
				if resp.StatusCode != expectedStatus {
					t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
				var errorResp model.ErrorResponse
				if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}
				if errorResp.Message != "account not found" {
					t.Errorf("Expected error message to be 'account not found', got '%s'", errorResp.Message)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transactionJSON, err := json.Marshal(tt.transaction)
			if err != nil {
				t.Fatalf("Failed to marshal transaction: %v", err)
			}

			req, err := http.NewRequest("POST", baseURL+"/transactions", bytes.NewBuffer(transactionJSON))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			req.Header.Set("Content-Type", "application/json")
			if tt.idempotencyKey != "" {
				req.Header.Set("X-idempotency-Key", tt.idempotencyKey)
			}
			resp, err := httpClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to create transaction: %v", err)
			}
			defer resp.Body.Close()

			tt.validate(t, tt.transaction, tt.expectedStatus, resp, err)

		})
	}
}

func transactionValidator(t *testing.T, transaction model.TransactionRequestBody, expectedStatus int, resp *http.Response) {
	t.Helper()
	if resp.StatusCode != expectedStatus {
		t.Fatalf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
	var transactionResponse model.TransactionResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&transactionResponse); err != nil {
		t.Fatalf("Failed to decode transaction response: %v", err)
	}
	if transactionResponse.AccountID != transaction.AccountID {
		t.Errorf("Expected account ID %s, got %s", transaction.AccountID, transactionResponse.AccountID)
	}
	if transactionResponse.Amount != transaction.Amount {
		t.Errorf("Expected amount %.2f, got %.2f", transaction.Amount, transactionResponse.Amount)
	}
	if transactionResponse.OperationID != transaction.OperationID {
		t.Errorf("Expected operation ID %d, got %d", transaction.OperationID, transactionResponse.OperationID)
	}
	if transactionResponse.TransactionID == "" {
		t.Error("Expected transaction ID to be set")
	}
}

func TestIntegration_SwaggerEndpoint(t *testing.T) {
	// Not quite sure if I need a full set of table tests for this..... for now, but this will do and proves the swagger is working
	resp, err := httpClient.Get(baseURL + "/swagger/")
	if err != nil {
		t.Fatalf("Failed to access swagger endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected swagger endpoint to return 200, got %d", resp.StatusCode)
	}
}

func TestIntegration_InvalidRoutes(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response, expectedStatus int, err error)
	}{
		{
			name:           "Non-existent endpoint",
			method:         "GET",
			url:            baseURL + "/nonexistent",
			expectedStatus: http.StatusNotFound,
			validate: func(t *testing.T, resp *http.Response, expectedStatus int, err error) {
				if resp.StatusCode != expectedStatus {
					t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			resp, err := httpClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			tt.validate(t, resp, tt.expectedStatus, err)
		})
	}
}

func Test_TransactionDischarge(t *testing.T) {
	accountReq := model.AccountRequestBody{
		DocumentNumber: bson.NewObjectID().Hex(),
	}

	accountJSON, _ := json.Marshal(accountReq)
	resp, err := httpClient.Post(baseURL+"/accounts", "application/json", bytes.NewBuffer(accountJSON))
	if err != nil {
		t.Fatalf("Failed to create test account: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create test account, status: %d", resp.StatusCode)
	}

	var account model.AccountResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		t.Fatalf("Failed to decode account response: %v", err)
	}

	accountID := account.AccountID

	tests := []struct {
		name             string
		debtTransactions []float64
		positiveAmount   float64
		expectedAmount   float64
		expectedStatus   int
		validate         func(t *testing.T, expectedStatus int, expectedResult float64, resp *http.Response, err error)
	}{
		{
			name:             "Example 2 from task",
			debtTransactions: []float64{-50.0, -23.5, -18.7},
			positiveAmount:   60.0,
			expectedAmount:   0,
			expectedStatus:   http.StatusCreated,
			validate: func(t *testing.T, expectedStatus int, expectedResult float64, resp *http.Response, err error) {
				if resp.StatusCode != expectedStatus {
					t.Fatalf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
				var transactionResponse model.TransactionResponseBody
				if err := json.NewDecoder(resp.Body).Decode(&transactionResponse); err != nil {
					t.Fatalf("Failed to decode transaction response: %v", err)
				}
				validateTransactionRecord(t, transactionResponse.TransactionID, expectedResult, mongoClient)
			},
		},
		{
			name:             "Example 3 from task",
			debtTransactions: []float64{},
			positiveAmount:   100.0,
			expectedAmount:   67.8,
			expectedStatus:   http.StatusCreated,
			validate: func(t *testing.T, expectedStatus int, expectedResult float64, resp *http.Response, err error) {
				if resp.StatusCode != expectedStatus {
					t.Fatalf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
				}
				var transactionResponse model.TransactionResponseBody
				if err := json.NewDecoder(resp.Body).Decode(&transactionResponse); err != nil {
					t.Fatalf("Failed to decode transaction response: %v", err)
				}
				validateTransactionRecord(t, transactionResponse.TransactionID, expectedResult, mongoClient)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			for _, debt := range tt.debtTransactions {
				transaction := model.TransactionRequestBody{
					AccountID:   accountID,
					OperationID: model.OperationTypePurchase,
					Amount:      debt,
				}
				transactionJSON, _ := json.Marshal(transaction)
				req, err := http.NewRequest("POST", baseURL+"/transactions", bytes.NewBuffer(transactionJSON))
				if err != nil {
					t.Fatalf("Failed to create request: %v", err)
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-idempotency-Key", uuid.NewString())
				_, err = httpClient.Do(req)
				if err != nil {
					t.Fatalf("Failed to create transaction: %v", err)
				}
			}
			positiveTransaction := model.TransactionRequestBody{
				AccountID:   accountID,
				OperationID: model.OperationTypePayment,
				Amount:      tt.positiveAmount,
			}
			positiveTransactionJSON, _ := json.Marshal(positiveTransaction)
			req, err := http.NewRequest("POST", baseURL+"/transactions", bytes.NewBuffer(positiveTransactionJSON))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-idempotency-Key", uuid.NewString())
			resp, err := httpClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to create transaction: %v", err)
			}
			defer resp.Body.Close()

			tt.validate(t, tt.expectedStatus, tt.expectedAmount, resp, err)
		})
	}
}

func validateTransactionRecord(t *testing.T, transactionID string, expectedBalanceAmount float64, client *mongo.Client) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	id, err := bson.ObjectIDFromHex(transactionID)
	if err != nil {
		t.Fatalf("Failed to convert transaction ID to ObjectID: %v", err)
	}
	tx := client.Database("pismo").Collection("transactions").FindOne(ctx, bson.M{"_id": id})
	var transaction model.Transaction
	if err := tx.Decode(&transaction); err != nil {
		t.Fatalf("Failed to decode transaction: %v", err)
	}
	if transaction.Balance != expectedBalanceAmount {
		t.Errorf("Expected balance %.2f, got %.2f", expectedBalanceAmount, transaction.Balance)
	}
}
