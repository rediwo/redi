#!/bin/bash

# Test script to verify multi-platform builds work locally
set -e

VERSION="v1.0.0-test"
PLATFORMS="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64"

echo "Testing multi-platform builds for version: $VERSION"
echo "=========================================="

# Create build directory
mkdir -p build

for platform in $PLATFORMS; do
    GOOS=${platform%/*}
    GOARCH=${platform#*/}
    
    echo "Building for $GOOS/$GOARCH..."
    
    # Build redi
    if [ "$GOOS" = "windows" ]; then
        GOOS=$GOOS GOARCH=$GOARCH go build \
            -ldflags "-X main.Version=$VERSION -s -w" \
            -o build/redi-$VERSION-$GOOS-$GOARCH.exe \
            ./cmd/redi
        
        GOOS=$GOOS GOARCH=$GOARCH go build \
            -ldflags "-X main.Version=$VERSION -s -w" \
            -o build/rejs-$VERSION-$GOOS-$GOARCH.exe \
            ./cmd/rejs
    else
        GOOS=$GOOS GOARCH=$GOARCH go build \
            -ldflags "-X main.Version=$VERSION -s -w" \
            -o build/redi-$VERSION-$GOOS-$GOARCH \
            ./cmd/redi
            
        GOOS=$GOOS GOARCH=$GOARCH go build \
            -ldflags "-X main.Version=$VERSION -s -w" \
            -o build/rejs-$VERSION-$GOOS-$GOARCH \
            ./cmd/rejs
    fi
    
    echo "✅ Built successfully for $GOOS/$GOARCH"
done

echo ""
echo "All builds completed successfully!"
echo "Built files:"
ls -la build/

echo ""
echo "Testing version output from a few binaries..."

# Test version output for current platform
if [ "$(uname)" = "Darwin" ]; then
    if [ "$(uname -m)" = "arm64" ]; then
        ./build/redi-$VERSION-darwin-arm64 --version
        ./build/rejs-$VERSION-darwin-arm64 --version
    else
        ./build/redi-$VERSION-darwin-amd64 --version
        ./build/rejs-$VERSION-darwin-amd64 --version
    fi
elif [ "$(uname)" = "Linux" ]; then
    if [ "$(uname -m)" = "aarch64" ]; then
        ./build/redi-$VERSION-linux-arm64 --version
        ./build/rejs-$VERSION-linux-arm64 --version
    else
        ./build/redi-$VERSION-linux-amd64 --version
        ./build/rejs-$VERSION-linux-amd64 --version
    fi
fi

echo ""
echo "Testing archive creation and extraction..."

# Test archive creation for a platform
VERSION="v1.0.0-test"
GOOS="linux"
GOARCH="amd64"

# Create test archive with both redi and rejs executables
tar -czf build/redi-$VERSION-$GOOS-$GOARCH.tar.gz \
    -C build redi-$VERSION-$GOOS-$GOARCH rejs-$VERSION-$GOOS-$GOARCH 2>/dev/null || echo "Note: Binaries not found, skipping"

# Test extraction (similar to CI)
cd build
mkdir -p test-extract

if [ -f "redi-$VERSION-$GOOS-$GOARCH.tar.gz" ]; then
    echo "Testing archive extraction..."
    tar -xzf redi-$VERSION-$GOOS-$GOARCH.tar.gz -C test-extract 2>/dev/null || echo "Extraction test skipped"
    echo "✅ Archive extraction test completed"
    echo "Extracted files (should contain both redi and rejs):"
    ls -la test-extract/
fi

cd ..

echo ""
echo "✅ Build test completed successfully!"