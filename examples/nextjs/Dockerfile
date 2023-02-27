FROM node:lts as dependencies
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

FROM node:lts
WORKDIR /app
COPY . .
COPY --from=dependencies /app/node_modules ./node_modules

EXPOSE 3000
