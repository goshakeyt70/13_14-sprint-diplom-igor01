# ЭТАП №1: Сборка и компиляция приложения (собираем код в один файл)

FROM golang:1.25-alpine AS builder

RUN apk add --no-cache build-base

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o sprint_13_14_scheduler_app .

# ЭТАП №2: Минимальная среда выполнения (Чистая консольная утилита)

FROM alpine:3.20.1

RUN apk --no-cache add libc6-compat && addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser
COPY --from=builder --chown=appuser:appgroup /app/sprint_13_14_scheduler_app .

COPY --from=builder --chown=appuser:appgroup /app/web ./web

ENV TODO_PORT=7540
ENV TODO_DBFILE=/data/scheduler.db

RUN mkdir -p /data && chown -R appuser:appgroup /data

VOLUME [ "/data" ]

EXPOSE 7540

USER appuser

ENTRYPOINT ["./sprint_13_14_scheduler_app"]