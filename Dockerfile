# syntax=docker/dockerfile:1

# ---- Build stage ----
# No more cgo dependency now that Oracle/godror is gone (see
# docs/adr/0013-mcp-server-transport.md and CONTEXT.md) - CGO_ENABLED=0 builds
# both binaries cleanly, so a plain Alpine Go image is enough here.
FROM golang:1.26-alpine AS build

ARG TARGETARCH

RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -trimpath -ldflags="-s -w" \
    -o /out/api ./cmd/api

# cmd/mcp-server is built here too - same image, same tag, same release - so
# k8s/mcp-server/00-deployment.yaml can just override `command` on this image
# instead of needing a second Dockerfile/build pipeline.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -trimpath -ldflags="-s -w" \
    -o /out/mcp-server ./cmd/mcp-server

# ---- Runtime stage ----
FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -g 10001 app \
    && adduser -D -H -u 10001 -G app -s /sbin/nologin app

WORKDIR /app

COPY --from=build /out/api ./api
COPY --from=build /out/mcp-server ./mcp-server
COPY migrations ./migrations

USER 10001:10001

EXPOSE 8080

ENTRYPOINT ["./api"]
