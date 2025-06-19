FROM golang:1.24.2-bullseye AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN apt-get update && apt-get install -y gcc libc6-dev

RUN CGO_ENABLED=1 GOOS=linux go build -o /app/relation-service ./cmd/server

FROM debian:bullseye-slim

WORKDIR /app

COPY --from=builder /app/relation-service .
COPY --from=builder /app/migrations ./migrations

EXPOSE 50054

CMD ["./relation-service"]