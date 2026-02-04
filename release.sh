#!/usr/bin/env bash
set -euo pipefail

# =========================================
# OpsLens Pulse Release Script — Production
# =========================================
# Builds Linux & Windows binaries, creates DEB packages,
# generates SHA256 checksums, and validates builds.

echo "📦 Starting OpsLens Pulse release..."

# ---------------------------
# Version handling
# ---------------------------
RAW_VERSION=${1:-${GITHUB_REF_NAME:-}}

if [[ -z "$RAW_VERSION" ]]; then
  echo "Usage: ./release.sh <version>"
  exit 1
fi

VERSION=${RAW_VERSION#v}

if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "❌ Invalid version format: $RAW_VERSION"
  exit 1
fi

echo "📦 Building release version: $VERSION"

# ---------------------------
# Constants
# ---------------------------
AGENT_APP="opslens-pulse-agent"
SERVER_APP="opslens-pulse-server"

ROOT_DIR="$(pwd)"
PACKAGE_DIR="$ROOT_DIR/package"
BUILD_DIR="$PACKAGE_DIR/build"
DIST_DIR="$PACKAGE_DIR/dist/releases"
DEB_BUILD="$BUILD_DIR/deb"

# ---------------------------
# Prepare directories
# ---------------------------
echo "🧹 Cleaning and preparing directories..."
rm -rf "$PACKAGE_DIR"
mkdir -p "$BUILD_DIR" "$DIST_DIR" "$DEB_BUILD"

# ---------------------------
# Go modules tidy
# ---------------------------
echo "🧹 Tidying Go modules..."
go mod tidy
(cd agent && go mod tidy)
(cd server && go mod tidy)

# ---------------------------
# Helper function: verify file exists
# ---------------------------
verify_file() {
  if [ ! -f "$1" ]; then
    echo "❌ Required file not found: $1"
    exit 1
  fi
}

# ---------------------------
# Build function
# ---------------------------
build_go_binary() {
  local target_dir="$1"
  local output_name="$2"
  local goos="$3"
  local goarch="$4"

  echo "🔨 Building $output_name for $goos/$goarch..."
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build -o "$target_dir/$output_name" || {
    echo "❌ Build failed for $output_name ($goos/$goarch)"
    exit 1
  }
  verify_file "$target_dir/$output_name"
  echo "✅ $output_name built"
}

# ---------------------------
# Linux builds
# ---------------------------
build_go_binary "$BUILD_DIR" "$AGENT_APP" linux amd64
build_go_binary "$BUILD_DIR" "$SERVER_APP" linux amd64

# ---------------------------
# DEB packaging function
# ---------------------------
create_deb_package() {
  local app_name="$1"
  local version="$2"
  local description="$3"

  local pkg_dir="$DEB_BUILD/$app_name"
  mkdir -p \
    "$pkg_dir/DEBIAN" \
    "$pkg_dir/usr/local/bin" \
    "$pkg_dir/etc/opslens-pulse" \
    "$pkg_dir/etc/systemd/system"

  cp "$BUILD_DIR/$app_name" "$pkg_dir/usr/local/bin/"

  # Control file
  cat > "$pkg_dir/DEBIAN/control" <<EOF
Package: $app_name
Version: $version
Section: utils
Priority: optional
Architecture: amd64
Maintainer: Vivek Bangare
Description: $description
EOF

  # Systemd service
  cat > "$pkg_dir/etc/systemd/system/$app_name.service" <<EOF
[Unit]
Description=$description
After=network.target

[Service]
ExecStart=/usr/local/bin/$app_name
Restart=always
RestartSec=5
User=root

[Install]
WantedBy=multi-user.target
EOF

  # Default config
  if [[ "$app_name" == "$AGENT_APP" ]]; then
    cat > "$pkg_dir/etc/opslens-pulse/agent-config.yaml" <<EOF
server:
  url: "http://localhost:9898"
  token: ""

agent:
  interval_seconds: 10
  self_upgrade: false
EOF
  else
    cat > "$pkg_dir/etc/opslens-pulse/server-config.yaml" <<EOF
listen_port: 9898
token: ""
EOF
  fi

  # Build DEB
  dpkg-deb --build "$pkg_dir"
  local deb_file="$DIST_DIR/${app_name}_${version}_amd64.deb"
  mv "$pkg_dir.deb" "$deb_file"
  verify_file "$deb_file"
  echo "✅ DEB package created: $deb_file"
}

# ---------------------------
# Create DEBs
# ---------------------------
create_deb_package "$AGENT_APP" "$VERSION" "OpsLens Pulse Host Monitoring Agent"
create_deb_package "$SERVER_APP" "$VERSION" "OpsLens Pulse Metrics Server"

# ---------------------------
# Windows builds
# ---------------------------
build_go_binary "$DIST_DIR" "${AGENT_APP}_${VERSION}_windows_amd64.exe" windows amd64
build_go_binary "$DIST_DIR" "${SERVER_APP}_${VERSION}_windows_amd64.exe" windows amd64

# ---------------------------
# Checksums
# ---------------------------
echo "🔐 Generating SHA256 checksums..."
cd "$DIST_DIR"
sha256sum * > SHA256SUMS.txt
verify_file "SHA256SUMS.txt"
echo "✅ Checksums generated: SHA256SUMS.txt"

# ---------------------------
# Summary
# ---------------------------
echo ""
echo "🎉 Release $RAW_VERSION built successfully!"
echo "📦 Artifacts in $DIST_DIR:"
ls -lh
