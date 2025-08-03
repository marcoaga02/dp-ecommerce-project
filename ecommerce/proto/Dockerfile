FROM golang:1.21-alpine AS builder

# install all needed dependencies
RUN apk add --no-cache \
        git \
        protobuf \
        protobuf-dev \
        build-base

WORKDIR /app

# Install protoc-gen-go and protoc-gen-go-grpc
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.29.0 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

ENV PATH="/go/bin:${PATH}"

COPY *.proto ./proto

# Script to compile all *.proto files
RUN echo '#!/bin/sh' > /usr/local/bin/compile-proto && \
    echo 'for protofile in "$@"; do' >> /usr/local/bin/compile-proto && \
    echo '    name=$(basename "$protofile" .proto)' >> /usr/local/bin/compile-proto && \
    echo '    mkdir -p "./$name"' >> /usr/local/bin/compile-proto && \
    echo '    protoc \' >> /usr/local/bin/compile-proto && \
    echo '      --go_out=paths=source_relative:"./$name" \' >> /usr/local/bin/compile-proto && \
    echo '      --go-grpc_out=paths=source_relative:"./$name" \' >> /usr/local/bin/compile-proto && \
    echo '      "$protofile"' >> /usr/local/bin/compile-proto && \
    echo 'done' >> /usr/local/bin/compile-proto && \
    chmod +x /usr/local/bin/compile-proto


#RUN echo '#!/bin/sh' > /usr/local/bin/compile-proto && \
#    echo 'protoc --go_out=paths=source_relative:./gen \' >> /usr/local/bin/compile-proto && \
#    echo '       --go-grpc_out=paths=source_relative:./gen "$@"' >> /usr/local/bin/compile-proto && \
#    chmod +x /usr/local/bin/compile-proto
    # "$@"' is a placeholder that will be substituted with the *.proto files when running the script
    # e.g. ./compile-proto *.proto
    
FROM alpine:latest

RUN apk add --no-cache protobuf

# Copy all compiled protoc-gen-go plugins necessary for the protoc execution, plus the compilation script
COPY --from=builder /go/bin/protoc-gen-go /usr/local/bin/
COPY --from=builder /go/bin/protoc-gen-go-grpc /usr/local/bin/
COPY --from=builder /usr/local/bin/compile-proto /usr/local/bin/

WORKDIR /app