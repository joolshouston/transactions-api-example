package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/joolshouston/pismo-technical-test/shared/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoDB struct {
	client *mongo.Client
}

func NewMongoDB(client *mongo.Client) *MongoDB {
	return &MongoDB{
		client: client,
	}
}

func (m *MongoDB) CreateAccount(ctx context.Context, documentID string) (*model.Account, error) {
	result, err := m.client.Database("pismo").Collection("accounts").InsertOne(ctx, model.Account{DocumentNumber: documentID})
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}
	return &model.Account{
		ID:             result.InsertedID.(bson.ObjectID),
		DocumentNumber: documentID,
	}, nil
}

func (m *MongoDB) GetAccountByID(ctx context.Context, accountID string) (*model.Account, error) {
	objectID, err := bson.ObjectIDFromHex(accountID)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID format: %w", err)
	}
	var acc model.Account
	err = m.client.Database("pismo").Collection("accounts").FindOne(ctx, bson.M{"_id": objectID}).Decode(&acc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	return &acc, nil
}

func (m *MongoDB) GetAccountByDocumentNumber(ctx context.Context, documentNumber string) (*model.Account, error) {
	var acc model.Account
	err := m.client.Database("pismo").Collection("accounts").
		FindOne(ctx, bson.M{"document_number": documentNumber}).Decode(&acc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account by document number: %w", err)
	}
	return &acc, nil
}

func (m *MongoDB) CreateTransaction(ctx context.Context, transaction model.Transaction) (*model.Transaction, error) {
	result, err := m.client.Database("pismo").Collection("transactions").InsertOne(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}
	transaction.ID = result.InsertedID.(bson.ObjectID)
	return &transaction, nil
}

func (m *MongoDB) FindTransactionByIdempotencyKey(ctx context.Context, idempotencyKey string) (*model.Transaction, error) {
	var tx model.Transaction
	err := m.client.Database("pismo").Collection("transactions").
		FindOne(ctx, bson.M{"idempotency_key": idempotencyKey}).Decode(&tx)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find transaction by idempotency key: %w", err)
	}
	return &tx, nil
}
