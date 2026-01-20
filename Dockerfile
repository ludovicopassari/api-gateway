FROM golang:1.26rc1-alpine3.23 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./pkg ./pkg

RUN go build -o /api-gateway ./cmd/gateway/main.go

FROM alpine:3.18
WORKDIR /root/
COPY --from=builder /api-gateway .
EXPOSE 8080
CMD ["./api-gateway"]
