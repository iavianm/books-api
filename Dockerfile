FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/api ./cmd/api

FROM alpine:3.22

RUN adduser -D -u 1000 appuser

WORKDIR /app
COPY --from=builder /app/api .

USER appuser
EXPOSE 8080

CMD ["./api"]
