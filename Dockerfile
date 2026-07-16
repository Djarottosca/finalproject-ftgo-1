# syntax=docker/dockerfile:1

# Shared build stage for all 4 binaries in this monorepo (core-server,
# core-worker, payment-server, notification-server) — one Dockerfile,
# binary picked via CMD_PATH build arg, to avoid 4 near-identical files.
FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG CMD_PATH
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/app ./${CMD_PATH}

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/app /app
ENTRYPOINT ["/app"]
