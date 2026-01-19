# Build stage
FROM golang:1.24-alpine AS builder

# Install git for version info and ca-certificates for HTTPS
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments for version injection
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE}" \
    -o chotko ./cmd/chotko

# Runtime stage
FROM alpine:3.20

# Labels
LABEL org.opencontainers.image.title="chotko" \
      org.opencontainers.image.description="Terminal UI for Zabbix monitoring" \
      org.opencontainers.image.source="https://github.com/harpchad/chotko" \
      org.opencontainers.image.licenses="MIT"

# Install ca-certificates for HTTPS connections to Zabbix
RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -u 1000 chotko

# Copy binary from builder
COPY --from=builder /app/chotko /usr/local/bin/chotko

# Use non-root user
USER chotko

# Config directory
VOLUME ["/home/chotko/.config/chotko"]

ENTRYPOINT ["chotko"]
