FROM golang:1.24-alpine3.21 AS builder


WORKDIR /app

COPY . .

RUN apk add --no-cache make
RUN make build_with_swag

FROM scratch

WORKDIR /app/

COPY --from=builder /app/build/. .
COPY .env.example .env
COPY --from=builder /app/docs /app/docs
RUN ["unset HTTP_PROXY HTTPS_PROXY"]

EXPOSE 8080

CMD ["/app/backend"]