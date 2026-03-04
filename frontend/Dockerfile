# Build stage
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -o /app/server ./cmd/server

# Runtime stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/server .
COPY static/ ./static/
EXPOSE 3000
ENTRYPOINT ["/app/server"]
