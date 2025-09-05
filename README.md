# Pismo Technical Test API

A small Go (Golang) HTTP API that manages accounts and financial transactions, persisting data in MongoDB. It exposes endpoints to:

- Create an account and fetch it by ID
- Create a transaction with idempotency support
- Explore API docs via Swagger UI

This repository includes Docker/Docker Compose for local development, a Makefile with helpful commands, Swagger/OpenAPI docs, and both unit and integration tests.

## Contents
- Features
- Prerequisites
- Quick Start (Docker Compose)
- Local Development (without Docker)
- Configuration
- API Endpoints
- API Documentation (Swagger)
- Testing
- Makefile Targets
- Troubleshooting

## Features
- Go 1.25 application (multi-stage Docker build)
- REST API built with chi router
- MongoDB persistence
- Idempotent transaction creation (via X-idempotency-Key)
- Swagger/OpenAPI documentation
- Unit and integration test suites

## Prerequisites
- Docker and Docker Compose (recommended for quickest start)
- Or, for local only:
  - Go 1.25+
  - MongoDB 6+

## Quick Start (Docker Compose)
This spins up both MongoDB and the application.

1. Build and run services:
   - make run-all-services
   - This will:
     - Build the app image
     - Start MongoDB and the API (default port 8080)

2. Verify the API is running:
   - curl http://localhost:8080/v1/swagger/
   - You should see Swagger UI load in the browser if you open that URL.

3. Stop all services when finished:
   - docker compose -f docker-compose-all.yml down -v

Notes:
- The application uses the environment variable MONGODB_URI and defaults to mongodb://root:password@localhost:27017/?authSource=admin when not set.
- The docker-compose-all.yml sets MONGODB_URI for the app to connect to the MongoDB service container.

## Local Development (without Docker)
Run MongoDB locally (with a root user) and start the app directly.

1. Start MongoDB (two options):
   - Using the provided compose just for MongoDB:
     - docker compose up -d
   - Or start MongoDB your own way ensuring it accepts:
     - username: root
     - password: password
     - port: 27017

2. Export MONGODB_URI (optional; the app defaults to the value below if unset):
   - export MONGODB_URI="mongodb://root:password@localhost:27017/?authSource=admin"

3. Run the app:
   - go run ./cmd/.

The server listens on http://localhost:8080

## Configuration
- MONGODB_URI
  - Description: MongoDB connection string used by the API.
  - Default: mongodb://root:password@localhost:27017/?authSource=admin
  - Example: export MONGODB_URI="mongodb://root:password@localhost:27017/?authSource=admin"

### Curl Examples
- Create account
  - curl -sS -X POST http://localhost:8080/v1/accounts -H "Content-Type: application/json" -d '{"document_number":"12345678900"}'

- Get account
  - curl -sS http://localhost:8080/v1/accounts/<account_id>

- Create transaction
  - curl -sS -X POST http://localhost:8080/v1/transactions \
    -H "Content-Type: application/json" \
    -H "X-idempotency-Key: demo-001" \
    -d '{"account_id":"<account_id>","operation_type_id":1,"amount":-100.50}'

## API Documentation (Swagger)
- Swagger UI: http://localhost:8080/v1/swagger/
- OpenAPI JSON: http://localhost:8080/v1/swagger/doc.json

Regenerating Swagger docs (if you change annotations):
- Requires swag CLI (github.com/swaggo/swag/cmd/swag)
- make gen-swagger
  - Note: If running from repo root, ensure your swag version supports the module structure. If needed, run: swag init -g ./cmd/main.go -d cmd,shared/model && swag fmt

## Testing
You can run unit tests and integration tests.

- Unit tests:
  - make test
  - Runs go test ./... with coverage and generates coverage.out

- Integration tests:
  - make test-integration
  - This will:
    - Build and start MongoDB + app via docker-compose-all.yml
    - Wait for the app to be ready
    - Run go tests tagged with integration under ./testing
    - Tear down containers

# My comments

I believe the requirements have been met with this current solution, there are clear places to refactor however, this is a good place to start.

I have gone with a project structure that should allow this to be expanded further, used a shared directory and swagger generation.

This gives scope to break out the API into multiple services, and also allows for a shared library of models and shared code.

There are some comments in the code around my thought process and what I would like to change in the future.

Testing will require some more thought, core components have been tested and integration tests added as well for a starting point.