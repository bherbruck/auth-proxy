# Node.js build stage for Vite UI
FROM node:20-alpine AS ui-builder

WORKDIR /app/ui

# Copy package files
COPY ui/package*.json ./

# Install dependencies (including devDependencies needed for build)
RUN npm ci

# Copy ui source
COPY ui/ ./

# Build the UI
RUN npm run build

# Go build stage
FROM golang:1.24.5-alpine AS go-builder

# Install git for go mod download
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Copy built UI static files to where Go embed expects them
COPY --from=ui-builder /app/ui/dist ./ui/dist

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o auth-proxy \
    ./cmd/authproxy

# Final stage - using distroless for minimal size with CA certs
FROM gcr.io/distroless/static-debian12:nonroot

# Copy timezone data and CA certificates from go-builder
COPY --from=go-builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=go-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from go-builder stage
COPY --from=go-builder /app/auth-proxy /auth-proxy

# Use non-root user
USER nonroot:nonroot

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/auth-proxy", "--help"] || exit 1

# Run the application
ENTRYPOINT ["/auth-proxy"]
