# Build stage
FROM golang:1.23-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o discord-bot-framework .

# Build individual bots
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o mtg-card-bot ./apps/mtg-card-bot
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o clippy-bot ./apps/clippy
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o music-bot ./apps/music

# Production stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata ffmpeg yt-dlp

# Create non-root user
RUN addgroup -g 1001 -S discord && \
    adduser -S -D -H -u 1001 -h /app -s /sbin/nologin -G discord -g discord discord

# Create app directory
WORKDIR /app

# Copy built binaries from builder
COPY --from=builder /app/discord-bot-framework /app/
COPY --from=builder /app/mtg-card-bot /app/
COPY --from=builder /app/clippy-bot /app/
COPY --from=builder /app/music-bot /app/

# Copy configuration templates
COPY --from=builder /app/config.example.json /app/
COPY --from=builder /app/env.example /app/

# Create necessary directories
RUN mkdir -p /app/data /app/cache /app/logs && \
    chown -R discord:discord /app

# Switch to non-root user
USER discord

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD ps aux | grep discord-bot-framework | grep -v grep || exit 1

# Default command
CMD ["./discord-bot-framework", "--bot", "all"]

# Metadata
LABEL maintainer="Discord Bot Framework" \
      version="2.0.0" \
      description="Modern Discord Bot Framework built with Go" \
      org.opencontainers.image.source="https://github.com/sawyer/discord-bot-framework" \
      org.opencontainers.image.documentation="https://github.com/sawyer/discord-bot-framework#readme" \
      org.opencontainers.image.licenses="MIT"