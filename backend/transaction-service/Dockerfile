FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY pkg/observability ./pkg/observability
COPY transaction-service/go.mod transaction-service/go.sum ./transaction-service/

WORKDIR /app/transaction-service
RUN go mod download && go mod verify

WORKDIR /app
COPY transaction-service ./transaction-service

WORKDIR /app/transaction-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/transaction-service/main .

EXPOSE 8085

CMD ["./main"]
