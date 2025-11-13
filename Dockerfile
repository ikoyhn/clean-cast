# syntax=docker/dockerfile:1

# Build arguments for multi-arch support
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

# Stage 1: Build the Go application
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.23.1-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
WORKDIR /app/cmd/app
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-w -s" -trimpath -o main .

# Stage 2: Runtime image with minimal footprint
FROM alpine:3.19

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Install runtime dependencies
RUN apk add --no-cache \
    ffmpeg \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

# Create necessary directories
RUN mkdir -p /app /config && \
    chown -R appuser:appuser /app /config

WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=appuser:appuser /app/cmd/app/main .

# Switch to non-root user
USER appuser

# Environment variables with defaults
ENV PORT=8080 \
    HOST=0.0.0.0 \
    AUDIO_DIR=/config/ \
    MIN_DURATION=5m

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]
