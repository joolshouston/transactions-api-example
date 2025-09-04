package controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/joolshouston/pismo-technical-test/cmd/services"
	"github.com/joolshouston/pismo-technical-test/shared/json_handler"
	"github.com/joolshouston/pismo-technical-test/shared/model"
)

type TransactionsController struct {
	service *services.TransactionService
	logger  *slog.Logger
}

func NewTransactionsController(service *services.TransactionService, logger *slog.Logger) *TransactionsController {
	return &TransactionsController{service: service, logger: logger}
}

// PostAccount 	 godoc
//
//	@Summary		Post a transaction
//	@Description	create a transaction
//	@Tags			transactions
//	@Param			transaction			body		model.TransactionRequestBody	true	"Transaction request body"
//	@Param			X-idempotency-Key	header		string							true	"Idempotency Key"
//	@Success		201					{object}	model.TransactionResponseBody
//	@Failure		400					{object}	model.ErrorResponse
//	@Failure		500					{object}	model.ErrorResponse
//	@Failure		409					{object}	model.ErrorResponse
//	@Failure		404					{object}	model.ErrorResponse
//	@Accept			json
//	@Produce		json
//	@Router			/transactions [post]
func (c *TransactionsController) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req model.TransactionRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json_handler.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	// validate x-idempotency-key header
	xIdempotencyKey := r.Header.Get("X-idempotency-Key")
	if xIdempotencyKey == "" {
		json_handler.WriteError(w, http.StatusBadRequest, "X-idempotency-Key header is required")
		return
	}
	if req.AccountID == "" || req.OperationID == 0 || req.Amount == 0 {
		json_handler.WriteError(w, http.StatusBadRequest, "account_id, operation_id and amount are required")
		return
	}

	transaction, err := c.service.CreateTransaction(r.Context(), req, xIdempotencyKey)
	if err != nil {
		json_handler.WriteError(w, http.StatusInternalServerError, "failed to create transaction")
		return
	}
	json_handler.WriteJSON(w, http.StatusCreated, transaction)
}
