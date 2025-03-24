.PHONY: build run test test-in-container clean

# Build the Docker image
build:
	docker compose build

# Run the application
run:
	docker compose up -d

# Run tests locally
test:
	go test -v ./...

# Run tests inside the container
test-in-container:
	docker compose run --rm app go test -v ./...

# Run integration tests inside the container
test-integration-in-container:
	docker compose run --rm -e TEST_INTEGRATION=true app go test -v ./integration_test.go

# Run database tests inside the container
test-db-in-container:
	docker compose run --rm -e TEST_DATABASE=true app go test -v ./pkg/db/...

# Run migrations
migrate:
	docker compose exec -it app sh -c 'goose -dir /app/internal/db/migrations postgres "$$PG_CONNECTION_STRING" up'

# Stop and remove containers
stop:
	docker compose down

# Clean up everything including volumes
clean:
	docker compose down -v
	rm -rf postgres-data

# Make a test API call
test-api:
	curl -X POST http://localhost:8080/chat -d "query=What is Go programming language?"

# Test file upload
test-upload:
	curl -X POST http://localhost:8080/upload -F "file=@README.md" -F "chat_id=test-session"

# Test chat with uploaded file
test-chat-with-file:
	curl -X POST http://localhost:8080/chat -H "Content-Type: application/json" -d '{"chat_id":"test-session","query":"What is this project about?"}'

# Test listing files
test-list-files:
	curl http://localhost:8080/files/test-session

# Test updating a file
test-update-file:
	curl -X PUT http://localhost:8080/files/test-session/README.md -F "file=@README.md"

# Test deleting a file
test-delete-file:
	curl -X DELETE http://localhost:8080/files/test-session/README.md

# Install swag
install-swag:
	go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger documentation
swagger: install-swag
	./scripts/generate-swagger.sh 