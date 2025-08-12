FROM golang:1.23-alpine AS build

ARG BUILD_MODE=dev

WORKDIR /app

COPY product-service ./product-service
COPY proto ./proto
COPY logger ./logger

WORKDIR /app/product-service

RUN if [ "$BUILD_MODE" = "prod" ]; then \
      sed -i '/replace.*=>/d' go.mod; \
    fi

RUN go mod download

# Compile statically for Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./main.go


FROM alpine:latest AS run

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=build /app/product-service/main /app/main

# Add the grpc_health_probe
ADD https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.16/grpc_health_probe-linux-amd64 /bin/grpc_health_probe
RUN chmod +x /bin/grpc_health_probe

CMD ["./main"]