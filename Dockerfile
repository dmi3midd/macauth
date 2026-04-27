# Builder stage

FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o macauth ./cmd/api

# Final stage

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/macauth .

COPY --from=builder /app/migrations ./migrations

# COPY --from=builder /app/assets ./assets

RUN mkdir -p keys storage

EXPOSE 2800

CMD ["./macauth"]
