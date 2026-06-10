FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /bin/api ./cmd/api

FROM alpine:3.21

RUN apk add --no-cache wget ca-certificates \
    && adduser -D -u 10001 app

WORKDIR /app
COPY --from=build /bin/api /app/api
COPY migrations /app/migrations

USER app

EXPOSE 8080
CMD ["/app/api"]
