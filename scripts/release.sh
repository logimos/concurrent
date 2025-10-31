#!/bin/bash

# Release script for concurrent library
# Usage: ./scripts/release.sh v1.0.0

set -e

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "Error: Version required"
    echo "Usage: ./scripts/release.sh v1.0.0"
    exit 1
fi

# Validate version format (basic check)
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Version must follow semantic versioning (e.g., v1.0.0)"
    exit 1
fi

echo "🚀 Preparing release $VERSION..."

# Check if there are uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo "⚠️  Warning: You have uncommitted changes."
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Check if tag already exists
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo "❌ Error: Tag $VERSION already exists"
    exit 1
fi

# Run tests
echo "🧪 Running tests..."
go test ./...

# Create tag
echo "📝 Creating tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"

# Show what will be pushed
echo ""
echo "✅ Tag created: $VERSION"
echo ""
echo "To publish, run:"
echo "  git push origin main"
echo "  git push origin $VERSION"
echo ""
echo "Or push everything at once:"
echo "  git push origin main --tags"
echo ""

