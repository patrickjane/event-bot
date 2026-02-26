# --- Stage 1: Builder ---
FROM golang:1.21-alpine AS builder

# Install tzdata and CA certificates
# We need these because 'scratch' is empty
RUN apk update && apk add --no-cache tzdata ca-certificates

# --- Stage 2: Final (Scratch) ---
FROM scratch

# 1. Copy the timezone data from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# 2. Copy SSL certificates so the bot can connect to Discord (HTTPS)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 3. Copy our static binary
ARG TARGETARCH
ARG VERSION

LABEL org.opencontainers.image.version=${VERSION}

COPY eventbot.linux-${TARGETARCH} /eventbot

# Set the timezone environment variable (Optional but helpful)
ENV TZ=Europe/Berlin

# Run the bot
ENTRYPOINT ["/eventbot"]