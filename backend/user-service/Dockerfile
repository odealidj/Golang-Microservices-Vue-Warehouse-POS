FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY pkg/observability ./pkg/observability
COPY user-service/go.mod user-service/go.sum ./user-service/

WORKDIR /app/user-service
RUN go mod download && go mod verify

WORKDIR /app
COPY user-service ./user-service

WORKDIR /app/user-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/user-service/main .

EXPOSE 8081

CMD ["./main"]
