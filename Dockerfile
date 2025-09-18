# --- build stage ---
FROM golang:1.24-alpine AS build
WORKDIR /app

# Cache dependecies
COPY go.mod go.sum ./
RUN go mod download

# All code
COPY . .

# Metadata params
ARG VERSION=dev
ARG COMMIT=dev
ARG BUILD_TIME=unknown

# Build binary file
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags "-s -w \
      -X github.com/berezovskyivalerii/tickersvc/internal/config.Version=${VERSION} \
      -X github.com/berezovskyivalerii/tickersvc/internal/config.Commit=${COMMIT} \
      -X github.com/berezovskyivalerii/tickersvc/internal/config.BuildTime=${BUILD_TIME}" \
    -o /bin/app ./cmd/api

# --- runtime stage ---
FROM alpine:3.20
RUN apk add --no-cache curl
ENV PORT=8080 GIN_MODE=release
EXPOSE 8080

COPY --from=build /bin/app /app

# Healthcheck /health
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD curl -fsS http://127.0.0.1:${PORT}/health || exit 1

ENTRYPOINT ["/app"]
