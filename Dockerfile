FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o user-service ./cmd/main.go
RUN mkdir -p /app/uploads

FROM gcr.io/distroless/static-debian12
WORKDIR /
COPY --from=builder /app/user-service /user-service
# Pre-create the uploads dir owned by nonroot so the named volume mounted
# there (empty on first run) inherits writable ownership instead of root's.
COPY --from=builder --chown=65532:65532 /app/uploads /data/uploads
USER nonroot:nonroot
EXPOSE 8087
ENTRYPOINT ["/user-service"]
