FROM golang:1.23-alpine AS build

ARG BUILD_MODE=dev

WORKDIR /app

COPY orchestrator ./orchestrator
COPY proto ./proto
COPY logger ./logger

WORKDIR /app/orchestrator

RUN if [ "$BUILD_MODE" = "prod" ]; then \
      sed -i '/replace.*=>/d' go.mod; \
    fi

RUN go mod download

# Compile statically for Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./main.go


FROM alpine:latest AS run

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=build /app/orchestrator/main /app/main
COPY --from=build /app/orchestrator/templates /app/orchestrator/templates

CMD ["./main"]