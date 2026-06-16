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
ENV PLAYWRIGHT_BROWSERS_PATH=/ms-playwright
RUN go run github.com/playwright-community/playwright-go/cmd/playwright install chromium webkit
COPY goserver/ ./
# Build the binary with CGO enabled
ENV CGO_ENABLED=1
RUN go build -ldflags="-w -s" -o dealscanner cmd/dealscanner/main.go

# Stage 3: Run the application
FROM debian:bookworm-slim
WORKDIR /app

# Install browser runtime dependencies for Playwright-backed scraping.
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      ca-certificates \
      chromium \
      fonts-liberation \
      gstreamer1.0-libav \
      gstreamer1.0-plugins-base \
      gstreamer1.0-plugins-good \
      libatomic1 \
      libatk-bridge2.0-0 \
      libatk1.0-0 \
      libavif15 \
      libcairo-gobject2 \
      libcairo2 \
      libdrm2 \
      libenchant-2-2 \
      libepoxy0 \
      libevent-2.1-7 \
      libflite1 \
      libfontconfig1 \
      libfreetype6 \
      libgbm1 \
      libgdk-pixbuf-2.0-0 \
      libgles2 \
      libglib2.0-0 \
      libgraphene-1.0-0 \
      libgstreamer-gl1.0-0 \
      libgstreamer-plugins-bad1.0-0 \
      libgstreamer-plugins-base1.0-0 \
      libgstreamer1.0-0 \
      libgtk-4-1 \
      libharfbuzz-icu0 \
      libharfbuzz0b \
      libhyphen0 \
      libicu72 \
      libjpeg62-turbo \
      liblcms2-2 \
      libmanette-0.2-0 \
      libopus0 \
      libpango-1.0-0 \
      libpangocairo-1.0-0 \
      libpng16-16 \
      libsecret-1-0 \
      libwayland-client0 \
      libwayland-egl1 \
      libwayland-server0 \
      libwebp7 \
      libwebpdemux2 \
      libwebpmux3 \
      libwoff1 \
      libx11-6 \
      libx264-164 \
      libxkbcommon0 \
      libxml2 \
      libxslt1.1 && \
    rm -rf /var/lib/apt/lists/*

# Environment variables for Playwright/Chromium.
ENV PUPPETEER_SKIP_DOWNLOAD=true
ENV PUPPETEER_EXECUTABLE_PATH=/usr/bin/chromium
ENV PLAYWRIGHT_BROWSERS_PATH=/ms-playwright

# Copy the built Go binary
COPY --from=backend-builder /app/goserver/dealscanner /app/dealscanner
COPY --from=backend-builder /root/.cache/ms-playwright-go /root/.cache/ms-playwright-go
COPY --from=backend-builder /ms-playwright /ms-playwright

# Copy the built frontend static assets
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist

# Cloud Run sets PORT at runtime. The Go config also accepts API_PORT for local runs.
EXPOSE 3000

# Set default env values
ENV API_HOST=0.0.0.0
ENV PORT=3000
ENV API_SECRET=dealscanner-dev-secret
ENV ENABLE_TESTING_API=true

# Run the app
CMD ["/app/dealscanner"]
