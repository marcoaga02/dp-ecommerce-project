FROM golang:1.23-alpine AS build

WORKDIR /app

COPY cart-service ./cart-service
COPY proto ./proto
COPY logger ./logger

WORKDIR /app/cart-service

RUN go mod download

# Compile statically for Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./main.go


FROM golang:1.23-alpine AS test

WORKDIR /app

COPY cart-service ./cart-service
COPY proto ./proto
COPY logger ./logger

WORKDIR /app/cart-service

RUN go mod download

CMD ["go", "test", "-v", "./..."]


FROM alpine:latest AS run

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=build /app/cart-service/main /app/main

# Add the grpc_health_probe
ADD https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.16/grpc_health_probe-linux-amd64 /bin/grpc_health_probe
RUN chmod +x /bin/grpc_health_probe

CMD ["./main"]