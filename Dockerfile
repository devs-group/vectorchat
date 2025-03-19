FROM golang:1.20-alpine AS builder

WORKDIR /app

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

# Copy the binary from the builder stage
COPY --from=builder /vectorchat .

# Command to run the executable
CMD ["./vectorchat"] 