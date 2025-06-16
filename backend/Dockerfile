# Build stage
FROM golang:1.23.3-alpine AS builder

WORKDIR /app

# Install build tools
RUN apk add --no-cache gcc musl-dev git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

WORKDIR /app

# Install CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/.env* .

# Expose the application port
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
