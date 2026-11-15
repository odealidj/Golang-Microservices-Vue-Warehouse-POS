FROM golang:1.24.3-alpine AS builder

WORKDIR /app

COPY pkg/observability ./pkg/observability
COPY warehouse-service/go.mod warehouse-service/go.sum ./warehouse-service/

WORKDIR /app/warehouse-service
RUN go mod download && go mod verify

WORKDIR /app
COPY warehouse-service ./warehouse-service

WORKDIR /app/warehouse-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/warehouse-service/main .

EXPOSE 8083

CMD ["./main"]
