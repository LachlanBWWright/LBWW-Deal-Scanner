# Stage 1: Build the Vue Frontend
FROM node:22-bookworm AS frontend-builder
WORKDIR /app/frontend
RUN corepack enable
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN printf "allowBuilds:\n  esbuild: true\n" > pnpm-workspace.yaml
RUN pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm run build

# Stage 2: Build the Go Backend
FROM golang:1.26-bookworm AS backend-builder
WORKDIR /app/goserver
# We need gcc for sqlite3 CGO compilation
RUN apt-get update && apt-get install -y gcc musl-dev
COPY goserver/go.mod goserver/go.sum ./
RUN go mod download
COPY goserver/ ./
# Build the binary with CGO enabled
ENV CGO_ENABLED=1
RUN go build -ldflags="-w -s" -o dealscanner cmd/dealscanner/main.go

# Stage 3: Run the application
FROM debian:bookworm-slim
WORKDIR /app

# Install chromium and dependencies for chromedp headless browser scraping
RUN apt-get update && \
    apt-get install -y --no-install-recommends chromium ca-certificates fonts-liberation && \
    rm -rf /var/lib/apt/lists/*

# Environment variables for Chromium/chromedp
ENV PUPPETEER_SKIP_DOWNLOAD=true
ENV PUPPETEER_EXECUTABLE_PATH=/usr/bin/chromium

# Copy the built Go binary
COPY --from=backend-builder /app/goserver/dealscanner /app/dealscanner

# Copy the built frontend static assets
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist

# Expose port
EXPOSE 3000

# Set default env values
ENV API_PORT=3000
ENV API_HOST=0.0.0.0
ENV DATABASE_URL=file:/app/db/dealscanner.db

# Run the app
CMD ["/app/dealscanner"]
