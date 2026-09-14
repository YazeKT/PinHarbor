# Build stage
FROM golang:1.26.6-trixie AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /usr/bin/google-maps-scraper

# Final stage - web runner only (no Playwright needed for API)
FROM debian:trixie-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/bin/google-maps-scraper /usr/bin/

WORKDIR /data
ENV RUN_MODE=web
ENV ADDR=0.0.0.0:8080
ENV DATA_FOLDER=/data

EXPOSE 8080
ENTRYPOINT ["google-maps-scraper"]
