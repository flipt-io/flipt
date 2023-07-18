FROM node:18-alpine3.16

WORKDIR /ui

COPY package.json .
COPY pnpm-lock.yaml .

RUN npm install -g pnpm
RUN pnpm install

EXPOSE 5173
