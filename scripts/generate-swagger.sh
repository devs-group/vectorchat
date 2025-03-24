#!/bin/bash
set -e

# Create docs directory if it doesn't exist
mkdir -p docs

# Generate swagger docs
swag init \
  --generalInfo cmd/main.go \
  --output docs \
  --parseInternal \
  --parseDependency \
  --parseDepth 2

echo "Swagger documentation generated successfully" 