# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS build
WORKDIR /src

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates && \
    adduser -D -u 10001 appuser
USER appuser

COPY --from=build /out/server /usr/local/bin/server

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/server"]
