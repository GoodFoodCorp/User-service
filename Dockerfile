FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o user-service ./cmd/main.go

FROM gcr.io/distroless/static-debian12
WORKDIR /
COPY --from=builder /app/user-service /user-service
USER nonroot:nonroot
EXPOSE 8087
ENTRYPOINT ["/user-service"]
