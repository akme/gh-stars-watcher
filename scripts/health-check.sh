#!/bin/bash

# Health check script for GitHub Stars Watcher Docker container
# This can be used with Docker's HEALTHCHECK instruction or external monitoring

set -e

# Check if the binary exists and is executable
if [ ! -x "/usr/local/bin/star-watcher" ]; then
    echo "❌ Binary not found or not executable"
    exit 1
fi

# Test that the binary can run and show help
if ! /usr/local/bin/star-watcher --help > /dev/null 2>&1; then
    echo "❌ Binary failed to execute"
    exit 1
fi

# Test that the state directory is accessible
if [ ! -d "/home/nonroot/.star-watcher" ]; then
    echo "❌ State directory not accessible"
    exit 1
fi

# Test write permissions in state directory
TEST_FILE="/home/nonroot/.star-watcher/.health-check"
if ! touch "$TEST_FILE" 2>/dev/null; then
    echo "❌ Cannot write to state directory"
    exit 1
fi
rm -f "$TEST_FILE"

echo "✅ Health check passed"
exit 0