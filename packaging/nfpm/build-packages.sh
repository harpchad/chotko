#!/bin/bash
set -e

VERSION="${1:-0.7.0}"
DIST_DIR="dist"

mkdir -p "$DIST_DIR"

# Build for each architecture
for GOARCH in amd64 arm64; do
	echo "Building chotko for linux/${GOARCH}..."

	# Build the binary
	CGO_ENABLED=0 GOOS=linux GOARCH=$GOARCH go build \
		-ldflags "-s -w -X main.version=${VERSION}" \
		-o chotko \
		./cmd/chotko

	# Build DEB package
	echo "Creating DEB package for ${GOARCH}..."
	GOARCH=$GOARCH nfpm package \
		--config packaging/nfpm/nfpm.yaml \
		--packager deb \
		--target "${DIST_DIR}/chotko_${VERSION}_${GOARCH}.deb"

	# Build RPM package
	echo "Creating RPM package for ${GOARCH}..."
	GOARCH=$GOARCH nfpm package \
		--config packaging/nfpm/nfpm.yaml \
		--packager rpm \
		--target "${DIST_DIR}/chotko-${VERSION}-1.${GOARCH}.rpm"

	# Cleanup
	rm -f chotko
done

echo "Packages created in ${DIST_DIR}/"
ls -la "${DIST_DIR}/"
