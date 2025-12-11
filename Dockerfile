# Етап 1: Збірка
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копіювати go.mod та go.sum для кешування
COPY go.mod go.sum ./
RUN go mod download

# Копіювати весь код
COPY . .

# Зібрати бота (без CGO для PostgreSQL не потрібен)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o movie-bot ./cmd/app

# Етап 2: Runtime
FROM alpine:latest

# Встановити CA сертифікати для HTTPS
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Створити непривілейованого користувача
RUN addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser && \
    chown -R appuser:appgroup /app

# Скопіювати бінарник
COPY --from=builder /app/movie-bot .
RUN chown appuser:appgroup /app/movie-bot

# Переключитися на непривілейованого користувача
USER appuser

# Запустити бота
CMD ["./movie-bot"]