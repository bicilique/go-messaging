# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Install dependencies first (better caching)
COPY app/go.mod app/go.sum ./
RUN go mod download

# Copy source code
COPY app/ .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o messaging-app ./cmd/main.go

# Run stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Create non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy binary from builder stage
COPY --from=builder /app/messaging-app .

# Change ownership and switch to non-root user
RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8080

CMD ["./messaging-app"]