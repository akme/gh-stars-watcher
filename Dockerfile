# Multi-architecture build arguments
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

FROM --platform=$BUILDPLATFORM golang:1.25.1-alpine AS builder

# Re-declare ARG variables for this stage
ARG TARGETARCH
ARG TARGETOS

RUN apk update && apk add --no-cache git ca-certificates tzdata file && update-ca-certificates

# Install UPX from Alpine package repository (simpler and more reliable)
RUN apk add --no-cache upx

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build for target architecture
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -ldflags='-w -s -extldflags "-static"' -a \
    -o /go/bin/star-watcher ./cmd/star-watcher

# Compress binary with UPX (skip for arm as it can be problematic)
RUN ARCH=${TARGETARCH:-amd64} && \
    if [ "${ARCH}" != "arm" ]; then \
        upx --overlay=strip --best /go/bin/star-watcher; \
    fi

# Create state directory structure and user files for scratch image
RUN mkdir -p /tmp/star-watcher/.star-watcher && \
    chown -R 65532:65532 /tmp/star-watcher && \
    \
    # Create minimal user files for scratch
    mkdir -p /tmp/etc && \
    echo 'nonroot:x:65532:65532:nonroot:/home/nonroot:/sbin/nologin' > /tmp/etc/passwd && \
    echo 'nonroot:x:65532:' > /tmp/etc/group && \
    \
    # Create directory structure for scratch
    mkdir -p /tmp/home/nonroot && \
    chown -R 65532:65532 /tmp/home/nonroot && \
    \
    # Create SSL certificate directory structure
    mkdir -p /tmp/etc/ssl/certs

# Use minimal scratch image for maximum security and size optimization
FROM scratch

# Copy CA certificates for HTTPS connections to GitHub API
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Copy minimal user management files
COPY --from=builder /tmp/etc/passwd /etc/passwd
COPY --from=builder /tmp/etc/group /etc/group

# Copy the application binary
COPY --from=builder /go/bin/star-watcher /usr/local/bin/star-watcher

# Copy state directory structure and home directory
COPY --from=builder --chown=65532:65532 /tmp/star-watcher /home/nonroot
COPY --from=builder --chown=65532:65532 /tmp/home/nonroot /home/nonroot

# Use nonroot user (numeric for scratch compatibility)
USER 65532:65532
WORKDIR /home/nonroot

ENTRYPOINT ["/usr/local/bin/star-watcher"]

# Alternative distroless version (commented out for reference):
# FROM gcr.io/distroless/static:nonroot
# COPY --from=builder /go/bin/star-watcher /usr/local/bin/star-watcher
# COPY --from=builder --chown=65532:65532 /tmp/star-watcher /home/nonroot
# USER nonroot:nonroot
# WORKDIR /home/nonroot
# ENTRYPOINT ["/usr/local/bin/star-watcher"]