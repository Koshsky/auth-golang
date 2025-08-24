FROM golang:1.24-alpine

RUN GRPC_HEALTH_PROBE_VERSION=v0.4.15 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

# Install build dependencies and prepare certs directory
RUN apk add --no-cache git libc6-compat wget curl netcat-openbsd && \
    mkdir -p /app/certs

WORKDIR /app
COPY auth-service/go.mod auth-service/go.sum ./auth-service/
WORKDIR /app/auth-service
RUN go mod download

WORKDIR /app
COPY auth-service ./auth-service
# Copy TLS certs from repo (if present)
COPY certs /app/certs

WORKDIR /app/auth-service

# Build the binary and make it executable
RUN CGO_ENABLED=0 go build -o auth-service ./cmd/auth-service/main.go && \
    chmod +x auth-service

# Create non-root user and set ownership
RUN addgroup -g 1001 -S appgroup && \
    adduser -S appuser -u 1001 -G appgroup && \
    chown -R appuser:appgroup /app

# Ports are configured at runtime via compose/env; EXPOSE is static for metadata
EXPOSE 50051

# Switch to non-root user
USER appuser

CMD ["./auth-service"]