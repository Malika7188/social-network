.PHONY: build run test clean migrate

# Build the application
build:
	go build -o bin/api cmd/api/main.go

# Run the application
run:
	go run cmd/api/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Create a new migration
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir pkg/db/migrations/sqlite -seq $$name

# Apply migrations
migrate-up:
	migrate -database sqlite3://data/social_network.db -path pkg/db/migrations/sqlite up

# Rollback migrations
migrate-down:
	migrate -database sqlite3://data/social_network.db -path pkg/db/migrations/sqlite down

# Create data directory
data:
	mkdir -p data 