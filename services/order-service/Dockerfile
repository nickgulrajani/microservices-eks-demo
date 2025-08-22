# Order Service Microservice
FROM golang:1.21-alpine AS builder

LABEL maintainer="DevOps Team"
LABEL service="order-service"
LABEL version="1.0.0"

WORKDIR /app

# Install git (required for go mod)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY src/ ./src/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main src/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Create non-root user
RUN adduser -D -s /bin/sh appuser
USER appuser

# Expose port
EXPOSE 3002

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3002/health || exit 1

CMD ["./main"]
