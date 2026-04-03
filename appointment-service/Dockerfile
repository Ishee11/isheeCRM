FROM golang:1.23.2-alpine AS builder
# Добавляем метку для удобства идентификации этапа
LABEL stage=gobuilder
# Отключаем использование C-библиотек для сборки статического бинарника
ENV CGO_ENABLED 0
ENV GOOS linux
# Указываем рабочую директорию для сборки
WORKDIR /build
# Копируем зависимости и скачиваем их в отдельный слой (оптимизация кэша)
COPY go.mod go.sum ./
RUN go mod download
# Копируем исходный код проекта
COPY . .
# Собираем бинарный файл приложения с оптимизацией размера
RUN go build -ldflags="-s -w" -o /app/main ./cmd/appointment-service

# Шаг 2: Используем минимальный образ Alpine для запуска приложения
FROM alpine:3.20
# Добавляем сертификаты для поддержки HTTPS
RUN apk add --no-cache ca-certificates wget
# Указываем рабочую директорию для приложения
WORKDIR /app
# Копируем собранный бинарный файл из builder-образа в конечный образ
COPY --from=builder /app/main .
# Копируем минимальный UI для проверки состояния сервиса после деплоя
COPY --from=builder /build/test.html ./test.html
# Открываем порт 8080
EXPOSE 8080
# Устанавливаем команду для запуска приложения
CMD ["/app/main"]
