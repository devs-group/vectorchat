FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install Air and required build tools
RUN apk add --no-cache gcc musl-dev git \
  && go install github.com/air-verse/air@latest \
  && go install github.com/pressly/goose/v3/cmd/goose@latest \
  && go install github.com/swaggo/swag/cmd/swag@latest

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /vectorchat ./cmd/main.go

# Use a small image for the final container
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Create uploads directory
RUN mkdir -p /root/uploads

# Copy the binary and other needed resources from the builder stage
COPY --from=builder /app/internal/db/migrations /migrations
COPY --from=builder /go/bin/goose /usr/local/bin/
COPY --from=builder /go/bin/air /usr/local/bin/
COPY --from=builder /vectorchat .

# Command to run the executable
CMD ["./vectorchat", "run"]
