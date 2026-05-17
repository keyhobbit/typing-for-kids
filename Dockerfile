# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Download dependencies first (layer-cache friendly)
COPY go.mod go.sum* ./
RUN go mod download

# Copy source and compile server binary (no GUI, no CGO needed for modernc sqlite)
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -tags '!gui' -ldflags="-s -w" -o kidtyping-server .

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM alpine:3.20

# ca-certificates for any future HTTPS outbound calls; tzdata for correct timestamps
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary and runtime assets
COPY --from=builder /app/kidtyping-server .
COPY static/    static/
COPY templates/ templates/

# SQLite database will be stored in /data (mounted as a volume)
ENV DB_PATH=/data/kidtyping.db

EXPOSE 11100

ENTRYPOINT ["./kidtyping-server"]
