FROM node:22

RUN mkdir /app
WORKDIR /app

ENV PUPPETEER_SKIP_DOWNLOAD=true
ENV PUPPETEER_EXECUTABLE_PATH=/usr/bin/chromium

# Puppeteer runtime dependencies. Debian's chromium package is available on
# both amd64 and arm64, which keeps this image compatible with OCI A1 Flex.
RUN apt-get update && \
    apt-get install -y --no-install-recommends chromium ca-certificates fonts-liberation && \
    rm -rf /var/lib/apt/lists/*

COPY server/package.json server/package-lock.json ./server/
COPY frontend/package.json frontend/package-lock.json ./frontend/
RUN npm --prefix server ci && npm --prefix frontend ci
COPY ./ ./

RUN npm --prefix server run build
RUN npm --prefix frontend run build

CMD ["npm", "--prefix", "server", "run", "start"]
