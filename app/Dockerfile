# Build stage
FROM golang:1.22-alpine@sha256:1699c10032ca2582ec89a24a1312d986a3f094aed3d5c1147b19880afe40e052 AS builder

# Set working directory
WORKDIR /src

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod verify

# Copy source code
COPY main.go ./

# Build static binary without CGO
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /bin/app main.go

# Production stage - Minimal Alpine image pinned by exact SHA digest
FROM alpine:3.20@sha256:d9e853e87e55526f6b2917df91a2115c36dd7c696a35be12163d44e6e2a4b6bc

# Create non-root user and group
RUN addgroup -g 10001 -S appgroup && \
    adduser -u 10001 -S appuser -G appgroup

# Run as non-root user
USER 10001:10001

# Copy application binary from builder
COPY --from=builder --chown=10001:10001 /bin/app /app

# Expose HTTP service port
EXPOSE 8080

# Entrypoint
ENTRYPOINT ["/app"]
