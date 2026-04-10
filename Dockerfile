FROM node:22

RUN mkdir /app
WORKDIR /app

COPY server/package.json server/package-lock.json ./server/
COPY frontend/package.json frontend/package-lock.json ./frontend/
RUN npm --prefix server install && npm --prefix frontend install
COPY ./ ./

#Puppeteer dependencies
RUN apt-get update && apt-get install gnupg wget -y && \
    wget -q -O- https://dl-ssl.google.com/linux/linux_signing_key.pub | gpg --dearmor > /etc/apt/trusted.gpg.d/google-archive.gpg && \
    sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list' && \
    apt-get update && \
    apt-get install google-chrome-stable -y --no-install-recommends && \
    rm -rf /var/lib/apt/lists/*

RUN npm --prefix server install @libsql/linux-x64-gnu

RUN npm --prefix server run build
RUN npm --prefix frontend run build

CMD ["npm", "--prefix", "server", "run", "start"]
