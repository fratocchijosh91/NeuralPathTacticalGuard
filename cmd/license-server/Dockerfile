FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /license-server ./cmd/license-server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -u 1000 appuser
RUN mkdir -p /app/data && chown -R appuser:appuser /app

COPY --from=builder /license-server /usr/local/bin/license-server

WORKDIR /app
USER appuser
EXPOSE 8080

ENTRYPOINT ["license-server"]
