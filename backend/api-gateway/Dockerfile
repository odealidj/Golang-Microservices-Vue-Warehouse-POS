FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY pkg/observability ./pkg/observability
COPY api-gateway/go.mod api-gateway/go.sum ./api-gateway/

WORKDIR /app/api-gateway
RUN go mod download && go mod verify

WORKDIR /app
COPY api-gateway ./api-gateway

WORKDIR /app/api-gateway
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/api-gateway/main .

EXPOSE 8080

CMD ["./main"]
