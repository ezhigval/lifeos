# syntax=docker/dockerfile:1

FROM node:22-alpine AS web
WORKDIR /web
# Vite's peak is the thing that OOM-kills a 2GB VM if the build is uncapped.
ENV NODE_OPTIONS=--max-old-space-size=768
COPY web/miniapp/package.json web/miniapp/package-lock.json ./
RUN npm install
COPY web/miniapp/ ./
RUN npm run build

FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# -p 1 keeps the compile inside the deploy script's memory cap.
RUN CGO_ENABLED=0 GOOS=linux GOMAXPROCS=1 go build -p 1 -ldflags="-s -w" -o /lifeos ./cmd/lifeos

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /lifeos /app/lifeos
COPY migrations /app/migrations
COPY --from=web /web/dist /app/web
COPY deployments/docker-entrypoint.sh /entrypoint.sh
RUN chmod 755 /entrypoint.sh

ENV LIFEOS_MIGRATIONS_DIR=/app/migrations
ENV LIFEOS_STATIC_DIR=/app/web
EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]
CMD ["serve"]
