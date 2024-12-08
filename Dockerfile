FROM node:22

RUN mkdir /app
WORKDIR /app


COPY package.json package-lock.json ./
RUN npm i
COPY ./ ./

#Puppeteer dependencies
RUN apt-get update && apt-get install gnupg wget -y && \
    wget -q -O- https://dl-ssl.google.com/linux/linux_signing_key.pub | gpg --dearmor > /etc/apt/trusted.gpg.d/google-archive.gpg && \
    sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list' && \
    apt-get update && \
    apt-get install google-chrome-stable -y --no-install-recommends && \
    rm -rf /var/lib/apt/lists/*

RUN npm i @libsql/linux-x64-gnu

RUN npx prisma generate
RUN npm run build

CMD ["npm", "run", "start"]