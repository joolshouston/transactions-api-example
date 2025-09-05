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
		json_handler.WriteError(w, &model.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid request body",
		})
		return
	}
	// validate x-idempotency-key header
	xIdempotencyKey := r.Header.Get("X-idempotency-Key")
	if xIdempotencyKey == "" {
		json_handler.WriteError(w, &model.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "X-idempotency-Key header is required",
		})
		return
	}
	// move to a better validator for the request bodies, we should take a set required attributes and validate, for now this will work
	if req.AccountID == "" || req.OperationID == 0 || req.Amount == 0 {
		json_handler.WriteError(w, &model.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "one or more of account_id, operation_type_id or amount is missing",
		})
		return
	}

	transaction, err := c.service.CreateTransaction(r.Context(), req, xIdempotencyKey)
	if err != nil {
		json_handler.WriteError(w, err)
		return
	}
	json_handler.WriteJSON(w, http.StatusCreated, transaction)
}
