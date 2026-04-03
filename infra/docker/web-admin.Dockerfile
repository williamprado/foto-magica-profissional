FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
COPY tsconfig.base.json tailwind.config.ts postcss.config.cjs ./
COPY apps ./apps
COPY packages ./packages
RUN npm ci
RUN npm run build -w apps/web-admin

FROM nginx:1.27-alpine
COPY infra/docker/nginx-spa.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/apps/web-admin/dist /usr/share/nginx/html

