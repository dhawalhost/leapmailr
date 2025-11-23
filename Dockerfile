# Multi-stage build for LeapMailr Backend
FROM golang:1.23-alpine AS builder

# Install build dependencies (sorted alphanumerically)
RUN apk add --no-cache ca-certificates git tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o leapmailr .

# Final stage - minimal runtime image
FROM alpine:latest

# Install runtime dependencies and create non-root user
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 leapmailr && \
    adduser -D -u 1000 -G leapmailr leapmailr

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/leapmailr .

# Copy necessary files
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Change ownership
RUN chown -R leapmailr:leapmailr /app

# Switch to non-root user
USER leapmailr

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./leapmailr"]
