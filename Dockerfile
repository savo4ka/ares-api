# Dockerfile для Ares API
# Multi-stage build для минимизации размера образа

# Stage 1: Builder
FROM golang:1.21-alpine AS builder

# Устанавливаем необходимые инструменты
RUN apk add --no-cache git ca-certificates tzdata

# Устанавливаем рабочую директорию
WORKDIR /build

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем весь код
COPY . .

# Собираем приложение
# CGO_ENABLED=0 для статической линковки
# -ldflags="-w -s" для уменьшения размера бинарника
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o ares-api \
    ./cmd/server/main.go

# Stage 2: Runtime
FROM alpine:3.19

# Устанавливаем CA сертификаты, timezone data и curl для healthcheck
RUN apk --no-cache add ca-certificates tzdata curl

# Создаём пользователя для запуска приложения
RUN addgroup -g 1000 ares && \
    adduser -D -u 1000 -G ares ares

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем бинарник из builder stage
COPY --from=builder /build/ares-api .

# Копируем миграции
COPY --chown=ares:ares migrations ./migrations

# Переключаемся на непривилегированного пользователя
USER ares

# Открываем порт
EXPOSE 8080

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Запускаем приложение
ENTRYPOINT ["/app/ares-api"]
