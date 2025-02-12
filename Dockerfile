FROM golang:1.23-bookworm AS builder

RUN mkdir /build

WORKDIR /build

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest

RUN swag init

RUN go build -o backend main.go

FROM docker.io/debian:12.9-slim

RUN mkdir /app

WORKDIR /app

COPY --from=builder /build/backend .

COPY .env.example .env

CMD ["/app/backend"]