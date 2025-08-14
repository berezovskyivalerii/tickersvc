FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN go build -o /out/app ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache curl
ENV PORT=8080
COPY --from=build /out/app /app
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --retries=3 CMD curl -fsS http://localhost:8080/health || exit 1
ENTRYPOINT ["/app"]
