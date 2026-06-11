FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /bin/api ./cmd/api

FROM alpine:3.21

RUN apk add --no-cache wget ca-certificates openssl \
    && adduser -D -u 10001 app \
    && mkdir -p /app/keys \
    && openssl genrsa -out /app/keys/private.pem 2048 \
    && openssl rsa -in /app/keys/private.pem -pubout -out /app/keys/public.pem \
    && chown -R app:app /app/keys \
    && chmod 600 /app/keys/private.pem

WORKDIR /app
COPY --from=build /bin/api /app/api
COPY migrations /app/migrations

USER app

EXPOSE 8080
ENV JWT_PRIVATE_KEY_PATH=/app/keys/private.pem
ENV JWT_PUBLIC_KEY_PATH=/app/keys/public.pem
CMD ["/app/api"]
