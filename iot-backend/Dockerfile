# ============================================================
# CHAIMAN IoT Backend — Dockerfile Hardened
# ============================================================

FROM golang:1.22-alpine AS builder

# Dépendances build
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copier go.mod/go.sum en premier (cache Docker)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copier le code source
COPY . .

# Build statique (pas de CGO = pas de dépendances libc)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o iot-backend \
    ./cmd/server

# ============================================================
# Image finale — distroless (pas de shell, pas de root)
# ============================================================
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copier le binaire
COPY --from=builder /app/iot-backend /app/iot-backend

# Copier les certificats CA système
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Non-root
USER nonroot:nonroot

# Healthcheck
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/app/iot-backend", "-healthcheck"]

EXPOSE 8080

ENTRYPOINT ["/app/iot-backend"]
