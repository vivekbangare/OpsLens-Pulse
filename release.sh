#!/usr/bin/env bash
set -e

echo "📦 Starting OpsLens Pulse release..."

# ---------------------------
# Version handling
# ---------------------------
RAW_VERSION=${1:-${GITHUB_REF_NAME}}

if [ -z "$RAW_VERSION" ]; then
  echo "Usage: ./release.sh <version>"
  echo "Example: ./release.sh v1.0.0"
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
AGENT_APP=opslens-pulse-agent
SERVER_APP=opslens-pulse-server

PACKAGE_DIR=package
BUILD=$PACKAGE_DIR/build
DIST=$PACKAGE_DIR/dist/releases

ROOT_DIR="$(pwd)"

# ---------------------------
# Clean & prepare dirs
# ---------------------------
rm -rf "$PACKAGE_DIR"
mkdir -p "$BUILD" "$DIST"

# ---------------------------
# Go modules tidy
# ---------------------------
echo "🧹 Tidying Go modules..."
go mod tidy
(cd agent && go mod tidy)
(cd server && go mod tidy)

# ==========================================================
# LINUX BUILDS (FIRST)
# ==========================================================
echo "🐧 Building Linux agent..."
cd agent
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
go build -o "$BUILD/$AGENT_APP"
cd "$ROOT_DIR"

[ -f "$BUILD/$AGENT_APP" ] || { echo "❌ Linux agent build failed"; exit 1; }

echo "🐧 Building Linux server..."
cd server
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
go build -o "$BUILD/$SERVER_APP"
cd "$ROOT_DIR"

[ -f "$BUILD/$SERVER_APP" ] || { echo "❌ Linux server build failed"; exit 1; }

# ==========================================================
# DEB PACKAGE — AGENT
# ==========================================================
echo "📦 Creating DEB package for agent..."

PKG_AGENT="$BUILD/deb/$AGENT_APP"
mkdir -p \
  "$PKG_AGENT/DEBIAN" \
  "$PKG_AGENT/usr/local/bin" \
  "$PKG_AGENT/etc/opslens-pulse" \
  "$PKG_AGENT/etc/systemd/system"

cp "$BUILD/$AGENT_APP" "$PKG_AGENT/usr/local/bin/"

cat > "$PKG_AGENT/DEBIAN/control" <<EOF
Package: $AGENT_APP
Version: $VERSION
Section: utils
Priority: optional
Architecture: amd64
Maintainer: Vivek Bangare
Description: OpsLens Pulse Host Monitoring Agent
EOF

cat > "$PKG_AGENT/etc/systemd/system/$AGENT_APP.service" <<EOF
[Unit]
Description=OpsLens Pulse Agent
After=network.target

[Service]
ExecStart=/usr/local/bin/$AGENT_APP
Restart=always
RestartSec=5
User=root

[Install]
WantedBy=multi-user.target
EOF

cat > "$PKG_AGENT/etc/opslens-pulse/agent-config.yaml" <<EOF
server:
  url: "http://localhost:9898"
  token: ""

agent:
  interval_seconds: 10
  self_upgrade: false
EOF

dpkg-deb --build "$PKG_AGENT"
mv "$BUILD/deb/$AGENT_APP.deb" \
   "$DIST/${AGENT_APP}_${VERSION}_amd64.deb"

echo "✅ Agent DEB created"

# ==========================================================
# DEB PACKAGE — SERVER
# ==========================================================
echo "📦 Creating DEB package for server..."

PKG_SERVER="$BUILD/deb/$SERVER_APP"
mkdir -p \
  "$PKG_SERVER/DEBIAN" \
  "$PKG_SERVER/usr/local/bin" \
  "$PKG_SERVER/etc/opslens-pulse" \
  "$PKG_SERVER/etc/systemd/system"

cp "$BUILD/$SERVER_APP" "$PKG_SERVER/usr/local/bin/"

cat > "$PKG_SERVER/DEBIAN/control" <<EOF
Package: $SERVER_APP
Version: $VERSION
Section: utils
Priority: optional
Architecture: amd64
Maintainer: Vivek Bangare
Description: OpsLens Pulse Metrics Server
EOF

cat > "$PKG_SERVER/etc/systemd/system/$SERVER_APP.service" <<EOF
[Unit]
Description=OpsLens Pulse Server
After=network.target

[Service]
ExecStart=/usr/local/bin/$SERVER_APP
Restart=always
User=root

[Install]
WantedBy=multi-user.target
EOF

cat > "$PKG_SERVER/etc/opslens-pulse/server-config.yaml" <<EOF
listen_port: 9898
token: ""
EOF

dpkg-deb --build "$PKG_SERVER"
mv "$BUILD/deb/$SERVER_APP.deb" \
   "$DIST/${SERVER_APP}_${VERSION}_amd64.deb"

echo "✅ Server DEB created"

# ==========================================================
# WINDOWS BUILDS (LAST — FIXED)
# ==========================================================
echo "🪟 Building Windows agent..."
cd agent
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
go build -o "../$DIST/${AGENT_APP}_${VERSION}_windows_amd64.exe"
cd "$ROOT_DIR"

[ -f "$DIST/${AGENT_APP}_${VERSION}_windows_amd64.exe" ] \
  || { echo "❌ Windows agent build failed"; exit 1; }

echo "✅ Windows agent EXE created"

echo "🪟 Building Windows server..."
cd server
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
go build -o "../$DIST/${SERVER_APP}_${VERSION}_windows_amd64.exe"
cd "$ROOT_DIR"

[ -f "$DIST/${SERVER_APP}_${VERSION}_windows_amd64.exe" ] \
  || { echo "❌ Windows server build failed"; exit 1; }

echo "✅ Windows server EXE created"

# ==========================================================
# CHECKSUMS
# ==========================================================
echo "🔐 Generating checksums..."
cd "$DIST"
sha256sum * > SHA256SUMS.txt

echo "✅ SHA256SUMS.txt created"

# ---------------------------
# Summary
# ---------------------------
echo ""
echo "🎉 Release $RAW_VERSION built successfully!"
echo "📦 Artifacts:"
ls -lh
