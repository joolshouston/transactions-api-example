package controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/joolshouston/pismo-technical-test/cmd/services"
	"github.com/joolshouston/pismo-technical-test/shared/json_handler"
	"github.com/joolshouston/pismo-technical-test/shared/model"
)

type AccountsController struct {
	service *services.AccountsService
	logger  *slog.Logger
}

func NewAccountsController(service *services.AccountsService, logger *slog.Logger) *AccountsController {
	return &AccountsController{service: service, logger: logger}
}

// PostAccount 	 godoc
//
//	@Summary		Create an account
//	@Description	create account with document number
//	@Tags			accounts
//	@Param			account	body		model.AccountRequestBody	true	"Account info"
//	@Success		201		{object}	model.AccountResponseBody
//	@Failure		400		{object}	model.ErrorResponse
//	@Failure		500		{object}	model.ErrorResponse
//	@Failure		409		{object}	model.ErrorResponse
//	@Failure		404		{object}	model.ErrorResponse
//	@Accept			json
//	@Produce		json
//	@Router			/accounts [post]
func (c *AccountsController) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var account model.AccountRequestBody
	if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
		json_handler.WriteError(w, &model.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid request body",
		})
		return
	}
	if account.DocumentNumber == "" {
		json_handler.WriteError(w, &model.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "document_number is required",
		})
		return
	}

	accountResponse, err := c.service.CreateAccount(r.Context(), account.DocumentNumber)
	if err != nil {
		json_handler.WriteError(w, err)
		return
	}

	json_handler.WriteJSON(w, http.StatusCreated, accountResponse)
}

// PostAccount 	 godoc
//
//	@Summary		Get a specific account by ID
//	@Description	get account by ID
//	@Tags			accounts
//	@Param			id	path		string	true	"Account ID"
//	@Success		201	{object}	model.AccountResponseBody
//	@Failure		400	{object}	model.ErrorResponse
//	@Failure		500	{object}	model.ErrorResponse
//	@Failure		409	{object}	model.ErrorResponse
//	@Failure		404	{object}	model.ErrorResponse
//	@Accept			json
//	@Produce		json
//	@Router			/accounts/{id} [get]
func (c *AccountsController) GetAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	id = strings.TrimSpace(id)
	if id == "" {
		json_handler.WriteError(w, &model.ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "account ID is required",
		})
		return
	}

	account, err := c.service.GetAccountByID(r.Context(), id)
	if err != nil {
		json_handler.WriteError(w, err)
		return
	}
	json_handler.WriteJSON(w, http.StatusOK, account)
}
