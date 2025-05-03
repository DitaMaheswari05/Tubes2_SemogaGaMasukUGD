# Stage 1: Build the Go app
FROM golang:1.24.2-alpine AS builder

WORKDIR /app

# Only copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Then copy the rest of the source code
COPY . .

# Build the binary
RUN go build -o main

# Expose the application port
EXPOSE 8080

# Start the app
CMD ["./main"]
