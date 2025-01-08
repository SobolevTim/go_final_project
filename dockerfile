# Используем официальный образ Go для сборки
FROM golang:1.23.0-alpine AS builder

# Установим рабочую директорию в контейнере
WORKDIR /app

# Установим необходимые зависимости для сборки
RUN apk add --no-cache gcc musl-dev

# Копируем go.mod и go.sum для скачивания зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем весь исходный код проекта в контейнер
COPY . .

# Компилируем приложение в бинарный файл
RUN go build -o myapp cmd/server/main.go

# Используем образ alpine с добавлением браузера и его зависимостей
FROM alpine:latest

# Создаем рабочую директорию для контейнера
WORKDIR /root/

# Копируем скомпилированное приложение из стадии сборки
COPY --from=builder /app/myapp .

# Копируем файлы
COPY web ./web
COPY migration ./migration

CMD ["./myapp"]