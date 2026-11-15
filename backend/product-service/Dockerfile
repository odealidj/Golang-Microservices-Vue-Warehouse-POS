FROM golang:1.24.3-alpine AS builder

WORKDIR /app

COPY pkg/observability ./pkg/observability
COPY product-service/go.mod product-service/go.sum ./product-service/

WORKDIR /app/product-service
RUN go mod download && go mod verify

WORKDIR /app
COPY product-service ./product-service

WORKDIR /app/product-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/product-service/main .

EXPOSE 8082

CMD ["./main"]
