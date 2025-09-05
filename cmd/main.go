package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/joolshouston/pismo-technical-test/cmd/controllers"
	"github.com/joolshouston/pismo-technical-test/cmd/services"
	"github.com/joolshouston/pismo-technical-test/shared/database"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Application struct {
	Logger *slog.Logger
}

//	@title			Pismo Technical Test API
//	@version		1.0
//	@description	This is a sample server for a financial transactions API.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8080
//	@BasePath	/v1

// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func main() {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(handler)
	ctx := context.Background()
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		logger.WarnContext(ctx, "MONGODB_URI not set, defaulting to mongodb://root:password@localhost:27017/?authSource=admin")
		mongoURI = "mongodb://root:password@localhost:27017/?authSource=admin"
	}
	mongoClient, err := setupMongoDB(ctx, mongoURI)
	if err != nil {
		logger.ErrorContext(ctx, "failed to set up MongoDB", "error", err)
		return
	}
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			logger.ErrorContext(ctx, "failed to disconnect MongoDB client", "error", err)
			return
		}
	}()
	logger.InfoContext(ctx, "MongoDB connected successfully")
	app := &Application{
		Logger: logger,
	}
	// Setup repository, service, and controller
	mongoRepo := database.NewMongoDB(mongoClient)
	accountService := services.NewAccountsService(mongoRepo, logger)
	accountController := controllers.NewAccountsController(accountService, logger)
	transactionService := services.NewTransactionService(mongoRepo, logger)
	transactionController := controllers.NewTransactionsController(transactionService, logger)

	// Routes
	r := app.routes(accountController, transactionController)

	// Start HTTP server
	addr := ":8080"
	logger.InfoContext(ctx, "starting server", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.ErrorContext(ctx, "server failed", "error", err)
	}
}

func setupMongoDB(ctx context.Context, mongoURI string) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, nil
}
