FROM golang:1.24.1-alpine

WORKDIR /app

# Install required system packages
RUN apk add --no-cache make gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application
COPY . .

# Create necessary directories
RUN mkdir -p data
RUN mkdir -p uploads

# Run the application in development mode
CMD ["go", "run", "cmd/api/main.go"]