gen-swagger:
	swag init -g ./main.go -d cmd,shared/model && swag fmt


run:
	docker compose up -d
	go run ./cmd/.

build-docker:
	docker build -t pismo-technical-test .

run-all-services:
	docker compose -f docker-compose-all.yml build --no-cache
	docker compose -f docker-compose-all.yml up -d --build
	
# Not sure if its my windows machine but having issues with stopping and removing the container
stop-pismo:
	docker compose -f docker-compose-all.yml stop
	docker compose -f docker-compose-all.yml rm -f pismo-app

test-integration:
	@echo "Building and starting application for integration tests..."
	docker compose -f docker-compose-all.yml build --no-cache
	docker compose -f docker-compose-all.yml up -d
	@echo "Waiting for services to be ready..."
	sleep 15
	@echo "Running integration tests..."
	go test -tags=integration -v ./testing/ || true
	@echo "Stopping integration test services..."
	docker compose -f docker-compose-all.yml down -v

test:
	go test -count=1 -v ./... -covermode=atomic -coverpkg=./... -coverprofile=coverage.out && go tool cover -func=coverage.out