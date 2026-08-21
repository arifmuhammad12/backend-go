# Multi-stage build — pola dari produksi (disederhanakan, tanpa DB)
# Stage 1: Build
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG APP_VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -o /app/server main.go

# Stage 2: Runtime (minimal)
FROM alpine:3.20
RUN addgroup -S appuser && adduser -S appuser -G appuser
RUN apk add --no-cache tzdata ca-certificates curl
ENV TZ=Asia/Jakarta
WORKDIR /app
COPY --from=builder --chown=appuser:appuser /app/server .
USER appuser
EXPOSE 4000
CMD ["/app/server"]
