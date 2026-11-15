FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY pkg/observability ./pkg/observability
COPY merchant-service/go.mod merchant-service/go.sum ./merchant-service/

WORKDIR /app/merchant-service
RUN go mod download && go mod verify

WORKDIR /app
COPY merchant-service ./merchant-service

WORKDIR /app/merchant-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/merchant-service/main .

EXPOSE 8084

CMD ["./main"]
