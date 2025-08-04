# Multi-stage build for smaller final image
FROM python:3.11-slim as builder

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install uv
RUN pip install uv

# Set work directory
WORKDIR /app

# Copy dependency files
COPY pyproject.toml ./
COPY apps/*/pyproject.toml ./apps/*/
COPY libs/*/pyproject.toml ./libs/*/

# Install dependencies
RUN uv venv /app/venv
ENV PATH="/app/venv/bin:$PATH"
RUN uv sync --all-extras

# Production stage
FROM python:3.11-slim

# Install runtime dependencies (FFmpeg for music bot)
RUN apt-get update && apt-get install -y \
    ffmpeg \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd --create-home --shell /bin/bash discord

# Copy virtual environment from builder
COPY --from=builder /app/venv /app/venv
ENV PATH="/app/venv/bin:$PATH"

# Set work directory
WORKDIR /app

# Copy application code
COPY --chown=discord:discord . .

# Create data directory for databases
RUN mkdir -p /app/data && chown discord:discord /app/data

# Switch to non-root user
USER discord

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Default command (can be overridden)
CMD ["python", "-m", "apps.music_bot.main"]