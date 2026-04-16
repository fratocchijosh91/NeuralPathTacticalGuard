FROM golang:1.22-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /license-server ./cmd/license-server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -u 1000 appuser

COPY --from=builder /license-server /usr/local/bin/license-server

USER appuser
EXPOSE 8080

ENTRYPOINT ["license-server"]
