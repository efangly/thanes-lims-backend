# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.26-alpine AS build

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags="-s -w" \
    -o /out/api ./cmd/api

# ---- Runtime stage ----
FROM alpine:3.21 AS runtime

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=build /out/api ./api
COPY migrations ./migrations

USER app

EXPOSE 8080

ENTRYPOINT ["./api"]
