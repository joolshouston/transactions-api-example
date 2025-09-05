package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joolshouston/pismo-technical-test/cmd/controllers"
	_ "github.com/joolshouston/pismo-technical-test/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func (app *Application) routes(accountController *controllers.AccountsController, transactionController *controllers.TransactionsController) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Timeout(60 * time.Second))

	mux.Route("/v1", func(r chi.Router) {
		// Account routes
		r.Post("/accounts", accountController.CreateAccount)
		r.Get("/accounts/{id}", accountController.GetAccount)

		// Transaction routes
		r.Post("/transactions", transactionController.CreateTransaction)

		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("http://localhost:8080/v1/swagger/doc.json")))
	})

	return mux
}
