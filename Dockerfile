# Используем официальный образ Go с конкретной версией для стабильности сборки
FROM golang:1.26.1 AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы go.mod и go.sum отдельно для кеширования зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости заранее
RUN go mod download

# Копируем весь исходный код приложения
COPY . .

# Собираем статически слинкованный бинарник для Alpine
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o subaggregator ./cmd/subaggregator

# ======= Этап 2: Минимальный образ для запуска приложения =======

# Используем минимальный стабильный образ Alpine Linux для финального контейнера
FROM alpine:3.21.3

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Добавляем корневые сертификаты для исходящих сетевых подключений
RUN apk add --no-cache ca-certificates

# Создаём непривилегированного пользователя и группу с фиксированными UID/GID
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Копируем готовый бинарник из стадии сборки
COPY --from=builder /app/subaggregator .

# Копируем папку с миграциями из стадии сборки в финальный образ
COPY --from=builder /app/migrations ./migrations

# Меняем владельца файлов на созданного пользователя
RUN chown -R appuser:appgroup /app

# Переключаемся на непривилегированного пользователя
USER appuser

EXPOSE 8080

# Устанавливаем команду по умолчанию для запуска контейнера
CMD ["./subaggregator"]
