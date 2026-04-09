# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Dependency layer (cache-friendly: changes only when go.mod/go.sum changes)
COPY go.mod go.sum ./
RUN go mod download

# Source
COPY . .

# Build all binaries with static linking (no libc dependency in final image)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/api    ./cmd/server/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/worker ./cmd/worker/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/mcp    ./cmd/mcp/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/seed   ./cmd/seed/

# ── API runtime ───────────────────────────────────────────────────────────────
FROM alpine:3.20 AS api

RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app
COPY --from=builder /bin/api ./api

EXPOSE 8080 8081

HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=5 \
  CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./api"]

# ── Worker runtime ────────────────────────────────────────────────────────────
FROM alpine:3.20 AS worker

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /bin/worker ./worker

CMD ["./worker"]

# ── MCP runtime (stdio — run locally, not as a container service) ─────────────
FROM alpine:3.20 AS mcp

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /bin/mcp ./mcp

CMD ["./mcp"]

# ── Seed runtime (one-shot demo data loader) ──────────────────────────────────
FROM alpine:3.20 AS seed

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /bin/seed ./seed

CMD ["./seed"]
