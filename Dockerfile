FROM golang:1.24.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /crypto-wallet-api cmd/main.go

FROM gcr.io/distroless/base

COPY --from=builder /crypto-wallet-api /crypto-wallet-api

EXPOSE 8080

CMD ["/crypto-wallet-api"]