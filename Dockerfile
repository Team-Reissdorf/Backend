FROM golang:1.24-alpine3.21 AS builder


WORKDIR /app

COPY . .

RUN apk add --no-cache make
RUN make build_with_swag


FROM scratch
WORKDIR /app/

COPY --from=builder /app/build/. /app/build/
COPY .env.example .env
COPY --from=builder /app/docs /app/docs

COPY ./rulesets/*.csv /rulesets/

EXPOSE 8080

CMD ["./build/backend"]