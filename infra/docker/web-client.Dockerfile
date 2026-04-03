FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
COPY tsconfig.base.json tailwind.config.ts postcss.config.cjs ./
COPY apps ./apps
COPY packages ./packages
RUN npm ci
ARG VITE_API_BASE_URL=https://fotomagica.wapainel.com.br/api
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
RUN npm run build -w apps/web-client

FROM nginx:1.27-alpine
COPY infra/docker/nginx-spa.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/apps/web-client/dist /usr/share/nginx/html

