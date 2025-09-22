# syntax=docker/dockerfile:1

# Install dependencies once and share across stages
FROM node:20-alpine AS base
WORKDIR /app
RUN apk add --no-cache git libc6-compat

COPY package*.json ./
RUN npm ci

# Development stage keeps dev dependencies and source files, runs Nuxt dev server
FROM base AS development
ENV NODE_ENV=development \
    NUXT_HOST=0.0.0.0 \
    NUXT_PORT=3000
COPY . .
EXPOSE 3000
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "3000"]

# Build stage produces production assets
FROM base AS build
ENV NODE_ENV=production
COPY . .
RUN npm run build

# Prune dev dependencies for production runtime
FROM build AS production-deps
RUN npm prune --omit=dev

# Production image contains only runtime deps and built output
FROM node:20-alpine AS production
WORKDIR /app
ENV NODE_ENV=production \
    NUXT_HOST=0.0.0.0 \
    NUXT_PORT=3000

# libc6-compat for node modules compiled against glibc
RUN apk add --no-cache libc6-compat

COPY --from=production-deps /app/node_modules ./node_modules
COPY --from=build /app/.output ./.output
COPY package*.json ./
EXPOSE 3000
CMD ["node", ".output/server/index.mjs"]
