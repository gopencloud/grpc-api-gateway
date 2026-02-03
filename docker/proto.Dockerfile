ARG ALPINE_VERSION=3.22
ARG GO_VERSION=1.25.2
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS base

ARG GID=1000
ARG UID=1000
RUN addgroup -g $GID -S usr && \
    adduser -u $UID -G usr -S usr -D

ARG EASYP_VERSION=v0.12.4
RUN apk update --no-cache && \
    apk add --no-cache bash git && \
    go install github.com/easyp-tech/easyp/cmd/easyp@${EASYP_VERSION} && \
    go clean -cache -testcache


FROM base AS gen

# Uncomment if you want to use local plugins for generation
#
#ARG PROTOC_GEN_GO_VERSION=v1.36.10
#ARG PROTOC_GEN_GO_GRPC_VERSION=v1.6.0
ARG PROTOC_GEN_JSONSCHEMA_VERSION=v0.8.0
RUN apk update --no-cache && \
    apk add --no-cache protobuf-dev && \
#    go install google.golang.org/protobuf/cmd/protoc-gen-go@${PROTOC_GEN_GO_VERSION} && \
#    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@${PROTOC_GEN_GO_GRPC_VERSION} && \
    go install github.com/pubg/protoc-gen-jsonschema@${PROTOC_GEN_JSONSCHEMA_VERSION} && \
    go clean -cache -testcache

USER usr

CMD ["sh", "-c", "easyp g"]


FROM base AS lint

USER usr

CMD ["sh", "-c", "easyp l"]