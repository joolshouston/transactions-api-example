package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Account struct {
	ID             bson.ObjectID `bson:"_id,omitempty"`
	DocumentNumber string        `bson:"document_number"`
}

// AccoundRequestBody model info
//
//	@Description	Account request body
//	@Description	Document number used to create an account
type AccountRequestBody struct {
	DocumentNumber string `json:"document_number"` // Document number
}

// AccountResponseBody model info
//
//	@Description	Account response body
//	@Description	ID and Document number of the created account
type AccountResponseBody struct {
	AccountID      string `json:"account_id"`
	DocumentNumber string `json:"document_number"`
}

// TransactionRequestBody model info
//
//	@Description	Transaction request body
//	@Description	Account ID, Operation type ID and Amount are required to create a transaction
type TransactionRequestBody struct {

	// required: true
	AccountID   string        `json:"account_id"`
	OperationID OperationType `json:"operation_type_id"`
	Amount      float64       `json:"amount"`
}

// TransactionResponseBody model info
//
//	@Description	Transaction response body
//	@Description	Transaction ID, Account ID, Operation type ID and Amount of the created transaction
type TransactionResponseBody struct {
	TransactionID string        `json:"transaction_id"`
	AccountID     string        `json:"account_id"`
	OperationID   OperationType `json:"operation_type_id"`
	Amount        float64       `json:"amount"`
}

type Transaction struct {
	ID             bson.ObjectID `bson:"_id,omitempty"`
	AccountID      string        `bson:"account_id"`
	OperationID    OperationType `bson:"operation_type_id"`
	Amount         float64       `bson:"amount"`
	EventDate      string        `bson:"event_date"`
	Balance        float64       `bson:"balance"`
	IdempotencyKey string        `bson:"idempotency_key"` // Idempotency Key this is to ensure idempotency of transactions, e.g., if the same request is sent multiple times, it will only be processed once
}

type OperationType int

const (
	OperationTypePurchase            OperationType = 1 //	@name	PURCHASE
	OperationTypeInstallmentPurchase OperationType = 2 //	@name	INSTALLMENT_PURCHASE
	OperationTypeWithdrawal          OperationType = 3 //	@name	WITHDRAWAL
	OperationTypePayment             OperationType = 4 //	@name	PAYMENT
)

func (ot OperationType) String() string {
	switch ot {
	case OperationTypePurchase:
		return "PURCHASE"
	case OperationTypeInstallmentPurchase:
		return "INSTALLMENT PURCHASE"
	case OperationTypeWithdrawal:
		return "WITHDRAWAL"
	case OperationTypePayment:
		return "PAYMENT"
	default:
		return "UNKNOWN"
	}
}

// ErrorResponse model info
//
//	@Description	Error response body
//	@Description	Message and Status code of the error
type ErrorResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}
