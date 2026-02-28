#!/usr/bin/env bash
set -euo pipefail

echo "📦 Starting OpsLens Pulse Agent release..."

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

AGENT_APP="opslens-pulse-agent"

ROOT_DIR="$(pwd)"
PACKAGE_DIR="$ROOT_DIR/package"
BUILD_DIR="$PACKAGE_DIR/build"
DIST_DIR="$PACKAGE_DIR/dist/releases"
DEB_BUILD="$BUILD_DIR/deb"

echo "🧹 Cleaning and preparing directories..."
rm -rf "$PACKAGE_DIR"
mkdir -p "$BUILD_DIR" "$DIST_DIR" "$DEB_BUILD"

echo "🧹 Tidying Go modules..."
go mod tidy
(cd agent && go mod tidy)

verify_file() {
  if [ ! -f "$1" ]; then
    echo "❌ Required file not found: $1"
    exit 1
  fi
}

# =====================================================
# 🔨 BUILD AGENT
# =====================================================

build_go_binary() {
  local source_dir="$1"
  local target_dir="$2"
  local output_name="$3"
  local goos="$4"
  local goarch="$5"

  echo "🔨 Building $output_name for $goos/$goarch..."

  (cd "$source_dir" && \
    GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 \
    go build -o "$target_dir/$output_name") || {
      echo "❌ Build failed for $output_name ($goos/$goarch)"
      exit 1
  }

  verify_file "$target_dir/$output_name"
  echo "✅ $output_name built"
}

# ---------------------------
# Linux build
# ---------------------------
build_go_binary "agent" "$BUILD_DIR" "$AGENT_APP" linux amd64

# =====================================================
# 📦 CREATE DEB (AGENT ONLY)
# =====================================================

create_agent_deb() {

  local pkg_dir="$DEB_BUILD/$AGENT_APP"

  mkdir -p \
    "$pkg_dir/DEBIAN" \
    "$pkg_dir/usr/local/bin" \
    "$pkg_dir/etc/opslens-pulse" \
    "$pkg_dir/etc/systemd/system"

  cp "$BUILD_DIR/$AGENT_APP" "$pkg_dir/usr/local/bin/"

  # Control file
  cat > "$pkg_dir/DEBIAN/control" <<EOF
Package: $AGENT_APP
Version: $VERSION
Section: utils
Priority: optional
Architecture: amd64
Maintainer: Vivek Bangare
Description: OpsLens Pulse Host Monitoring Agent
EOF

  # Systemd service
  cat > "$pkg_dir/etc/systemd/system/$AGENT_APP.service" <<EOF
[Unit]
Description=OpsLens Pulse Host Monitoring Agent
After=network.target

[Service]
ExecStart=/usr/local/bin/$AGENT_APP
Restart=always
RestartSec=5
User=root

[Install]
WantedBy=multi-user.target
EOF

  # =============================================
  # 🔥 UPDATED AGENT CONFIG (YOUR NEW CONFIG)
  # =============================================

  cat > "$pkg_dir/etc/opslens-pulse/agent-config.yaml" <<EOF
version: "1"

server:
  url: http://server:9898
  api_key: ""
  insecure_skip_verify: false

agent:
  interval_seconds: 5
  self_upgrade: false
  log_collection_interval_seconds: 5

  monitored_services:
    - nginx
    - ssh

  logs:
    - name: random_app_log
      type: file
      path: /logs/random.log

    - name: system_logs
      type: directory
      path: /var/log
      include:
        - syslog
        - auth.log
        - messages
      exclude:
        - "*.gz"
        - "*.old"

    - name: journal_logs
      type: journal

    - name: kernel_logs
      type: dmesg

tags:
  env: preprod
  region: us-west-1
  role: go-agent
  name: agent-config1
  owner: team-a
  product: opslens-pulse
  tier: backend
EOF

  dpkg-deb --build "$pkg_dir"
  local deb_file="$DIST_DIR/${AGENT_APP}_${VERSION}_amd64.deb"
  mv "$pkg_dir.deb" "$deb_file"

  verify_file "$deb_file"
  echo "✅ DEB package created: $deb_file"
}

create_agent_deb

# ---------------------------
# Windows build
# ---------------------------
build_go_binary "agent" "$DIST_DIR" "${AGENT_APP}_${VERSION}_windows_amd64.exe" windows amd64

# ---------------------------
# Checksums
# ---------------------------
echo "🔐 Generating SHA256 checksums..."
cd "$DIST_DIR"
sha256sum * > SHA256SUMS.txt
verify_file "SHA256SUMS.txt"

echo ""
echo "🎉 Agent Release $RAW_VERSION built successfully!"
echo "📦 Artifacts in $DIST_DIR:"
ls -lh