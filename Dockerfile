FROM golang:1.23-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Компилируем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o forza-bot ./cmd/bot/main.go

# Финальный образ
FROM alpine:3.18

WORKDIR /app

# Устанавливаем зависимости с переключением на CDN зеркала в случае ошибок
RUN apk update && \
    apk --no-cache add ca-certificates tzdata

# Копируем скомпилированное приложение
COPY --from=builder /app/forza-bot /app/
COPY --from=builder /app/configs /app/configs

# Добавляем пользователя для запуска приложения
RUN adduser -D -g '' appuser
USER appuser

# Запускаем приложение
CMD ["/app/forza-bot"]