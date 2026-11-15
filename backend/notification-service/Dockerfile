FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY pkg/observability ./pkg/observability
COPY notification-service/go.mod notification-service/go.sum ./notification-service/

WORKDIR /app/notification-service
RUN go mod download

WORKDIR /app
COPY notification-service ./notification-service

WORKDIR /app/notification-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/notification-service/main .

EXPOSE 8086

CMD ["./main"]
