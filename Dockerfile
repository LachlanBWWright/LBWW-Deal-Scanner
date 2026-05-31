FROM node:22

RUN mkdir /app
WORKDIR /app

RUN corepack enable

ENV PUPPETEER_SKIP_DOWNLOAD=true
ENV PUPPETEER_EXECUTABLE_PATH=/usr/bin/chromium

# Puppeteer runtime dependencies. Debian's chromium package is available on
# both amd64 and arm64, which keeps this image compatible with OCI A1 Flex.
RUN apt-get update && \
    apt-get install -y --no-install-recommends chromium ca-certificates fonts-liberation && \
    rm -rf /var/lib/apt/lists/*

COPY server/package.json server/pnpm-lock.yaml ./server/
COPY frontend/package.json frontend/pnpm-lock.yaml ./frontend/
RUN pnpm --dir server install --frozen-lockfile && pnpm --dir frontend install --frozen-lockfile
COPY ./ ./

RUN pnpm --dir server run build
RUN pnpm --dir frontend run build

CMD ["pnpm", "--dir", "server", "run", "start"]
