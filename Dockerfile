FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Install goose
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /vectorchat ./cmd/main.go

# Use a small image for the final container
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Create uploads directory
RUN mkdir -p /root/uploads

# Copy the binary and other needed resources from the builder stage
COPY --from=builder /app/pkg/db/migrations /migrations
COPY --from=builder /go/bin/goose /usr/local/bin/
COPY --from=builder /vectorchat .

# Command to run the executable
CMD ["./vectorchat"] 
