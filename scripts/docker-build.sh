#!/bin/bash

# Multi-architecture Docker build script for GitHub Stars Watcher
set -e

# Configuration
IMAGE_NAME="ghcr.io/akme/gh-stars-watcher"
PLATFORMS="linux/amd64,linux/arm64,linux/arm/v7"

# Check if buildx is available
if ! docker buildx version > /dev/null 2>&1; then
    echo "❌ Docker buildx is required for multi-architecture builds"
    echo "Please install Docker Desktop or enable buildx"
    exit 1
fi

# Create builder instance if it doesn't exist
BUILDER_NAME="multiarch-builder"
if ! docker buildx ls | grep -q "$BUILDER_NAME"; then
    echo "🔨 Creating buildx instance: $BUILDER_NAME"
    docker buildx create --name "$BUILDER_NAME" --driver docker-container --use
    docker buildx inspect --bootstrap
fi

# Use the builder
docker buildx use "$BUILDER_NAME"

# Function to build and push
build_and_push() {
    local tag=$1
    local push_flag=$2
    
    echo "🚀 Building multi-architecture image: $IMAGE_NAME:$tag"
    echo "📋 Platforms: $PLATFORMS"
    
    # Get current timestamp and git info for labels
    local build_date
    local git_sha
    local git_ref
    build_date=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    git_sha=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    git_ref=$(git symbolic-ref -q --short HEAD 2>/dev/null || git describe --tags --exact-match 2>/dev/null || echo "unknown")
    
    docker buildx build \
        --platform "$PLATFORMS" \
        --tag "$IMAGE_NAME:$tag" \
        --label "org.opencontainers.image.created=$build_date" \
        --label "org.opencontainers.image.revision=$git_sha" \
        --label "org.opencontainers.image.ref.name=$git_ref" \
        --label "org.opencontainers.image.version=$tag" \
        $push_flag \
        --progress=plain \
        .
}

# Parse command line arguments
case "${1:-local}" in
    "push")
        echo "🔄 Building and pushing to registry..."
        build_and_push "latest" "--push"
        
        # If there's a version tag, also push that
        if [ -n "$2" ]; then
            build_and_push "$2" "--push"
        fi
        ;;
    "local")
        echo "🏠 Building for local use..."
        # For local builds, use current platform only to enable --load
        ORIGINAL_PLATFORMS="$PLATFORMS"
        PLATFORMS="linux/$(docker info --format '{{.Architecture}}')"
        build_and_push "latest" "--load"
        PLATFORMS="$ORIGINAL_PLATFORMS"
        ;;
    "test")
        echo "🧪 Test build (no output)..."
        # Get current timestamp and git info for labels
        build_date=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
        git_sha=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
        git_ref=$(git symbolic-ref -q --short HEAD 2>/dev/null || git describe --tags --exact-match 2>/dev/null || echo "unknown")
        
        docker buildx build \
            --platform "$PLATFORMS" \
            --label "org.opencontainers.image.created=$build_date" \
            --label "org.opencontainers.image.revision=$git_sha" \
            --label "org.opencontainers.image.ref.name=$git_ref" \
            --progress=plain \
            .
        ;;
    *)
        echo "Usage: $0 [local|push|test] [version]"
        echo ""
        echo "Commands:"
        echo "  local  - Build for local use (default)"
        echo "  push   - Build and push to registry"
        echo "  test   - Test build without creating image"
        echo ""
        echo "Examples:"
        echo "  $0 local"
        echo "  $0 push"
        echo "  $0 push v1.0.0"
        exit 1
        ;;
esac

echo "✅ Build completed successfully!"