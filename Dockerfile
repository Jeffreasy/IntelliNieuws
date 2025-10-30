# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies (cached layer if go.mod/go.sum unchanged)
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./cmd/api

# Final stage - minimal distroless-like image
FROM alpine:latest

# Install runtime dependencies and curl for healthcheck
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    curl

# Create non-root user
RUN adduser -D -s /bin/sh appuser

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy migrations directory for database setup
COPY --from=builder /app/migrations ./migrations

# Change ownership to non-root user
RUN chown -R appuser:appuser /root/

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Improved health check with curl and better timing
HEALTHCHECK --interval=30s --timeout=5s --start-period=40s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Run the binary
CMD ["./main"]