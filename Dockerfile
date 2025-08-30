# Образ для билда
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Копируем проект
COPY . .

# Загружаем зависимости
RUN go mod download

# Собираем исполняемый файл
RUN go build -o main ./cmd/parser

# Минимальный образ для запуска
FROM alpine:3.22
WORKDIR /app

# Копируем из образа билда скомпилированный файл
COPY --from=builder /app/main .

