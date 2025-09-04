gen-swagger:
	swag init -g ./main.go -d cmd,shared/model && swag fmt

run:
	go run cmd/api/main.go

build-docker:
	docker build -t pismo-technical-test .

run-all-services:
	docker compose -f docker-compose-all.yml build --no-cache
	docker compose -f docker-compose-all.yml up -d --build
	
# Not sure if its my windows machine but having issues with stopping and removing the container
stop-pismo:
	docker compose -f docker-compose-all.yml stop
	docker compose -f docker-compose-all.yml rm -f pismo-app

test:
	go test -count=1 -v ./... -covermode=atomic -coverpkg=./... -coverprofile=coverage.out && go tool cover -func=coverage.out