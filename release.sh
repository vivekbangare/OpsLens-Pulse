#!/usr/bin/env bash
set -e

# ---------------------------
# Version handling
# ---------------------------
RAW_VERSION=${1:-${GITHUB_REF_NAME}}

if [ -z "$RAW_VERSION" ]; then
  echo "Usage: ./release.sh <version>"
  echo "Example: ./release.sh v1.0.0"
  exit 1
fi

# Strip leading 'v' for Debian packages
VERSION=${RAW_VERSION#v}

# Validate version format
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "❌ Invalid version format: $RAW_VERSION"
  echo "Expected: vX.Y.Z or X.Y.Z"
  exit 1
fi

AGENT_APP=opslens-pulse-agent
SERVER_APP=opslens-pulse-server
PACKAGE_DIR=package
BUILD=$PACKAGE_DIR/build
DIST=$PACKAGE_DIR/dist/releases

# Clean previous outputs
rm -rf $PACKAGE_DIR
mkdir -p $BUILD $DIST

# ---------------------------
# Go modules tidy
# ---------------------------
echo "🧹 Tidying Go modules..."
go mod tidy
(cd agent && go mod tidy)
(cd server && go mod tidy)

# ---------------------------
# Build Linux binaries
# ---------------------------
echo "🔧 Building Linux agent..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
go build -o $BUILD/$AGENT_APP ./agent

[ -f "$BUILD/$AGENT_APP" ] || { echo "❌ Linux agent build failed!"; exit 1; }

echo "🔧 Building Linux server..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
go build -o $BUILD/$SERVER_APP ./server

[ -f "$BUILD/$SERVER_APP" ] || { echo "❌ Linux server build failed!"; exit 1; }

# ---------------------------
# Create DEB package for agent
# ---------------------------
echo "📦 Creating DEB package for agent..."
PKG=$BUILD/deb/$AGENT_APP
mkdir -p \
  $PKG/DEBIAN \
  $PKG/usr/local/bin \
  $PKG/etc/opslens-pulse \
  $PKG/etc/systemd/system

cp $BUILD/$AGENT_APP $PKG/usr/local/bin/

cat > $PKG/DEBIAN/control <<EOF
Package: $AGENT_APP
Version: $VERSION
Section: utils
Priority: optional
Architecture: amd64
Maintainer: Vivek Bangare
Description: OpsLens Pulse Host Monitoring Agent
EOF

cat > $PKG/etc/systemd/system/$AGENT_APP.service <<EOF
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

cat > $PKG/etc/opslens-pulse/agent-config.yaml <<EOF
server:
  url: "http://localhost:9898"
  token: ""

agent:
  interval_seconds: 10
  self_upgrade: false
EOF

dpkg-deb --build $PKG
mv $BUILD/deb/$AGENT_APP.deb $DIST/${AGENT_APP}_${VERSION}_amd64.deb
echo "✅ DEB package for agent created"

# ---------------------------
# Build Windows agent
# ---------------------------
echo "🪟 Building Windows agent..."
(cd agent && GOOS=windows GOARCH=amd64 \
go build -o ../../$DIST/${AGENT_APP}_${RAW_VERSION}_windows_amd64.exe)

WIN_INSTALL_SCRIPT="$DIST/install-windows-agent.ps1"
cat > "$WIN_INSTALL_SCRIPT" <<'EOF'
param(
    [string]$ConfigPath = ""
)

$SourceDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$AgentExe = Join-Path $SourceDir "opslens-pulse-agent.exe"
$ConfigSrc = Join-Path $SourceDir "WindowsConfig\agent-config.yaml"

if ($ConfigPath -eq "") {
    $ConfigDestDir = "C:\ProgramData\OpsLens-Pulse"
} else {
    $ConfigDestDir = $ConfigPath
}

New-Item -ItemType Directory -Path $ConfigDestDir -Force | Out-Null

Copy-Item $ConfigSrc (Join-Path $ConfigDestDir "agent-config.yaml") -Force
Copy-Item $AgentExe (Join-Path $ConfigDestDir "opslens-pulse-agent.exe") -Force

Write-Host "✅ OpsLens agent installed at $ConfigDestDir"
EOF
echo "✅ Windows agent install script created at $WIN_INSTALL_SCRIPT"

# ---------------------------
# Build Windows server
# ---------------------------
echo "🪟 Building Windows server..."
(cd server && GOOS=windows GOARCH=amd64 \
go build -o ../../$DIST/${SERVER_APP}_${RAW_VERSION}_windows_amd64.exe)

# ---------------------------
# Create DEB package for server
# ---------------------------
echo "📦 Creating DEB package for server..."
PKG_SRV=$BUILD/deb/$SERVER_APP

mkdir -p \
  $PKG_SRV/DEBIAN \
  $PKG_SRV/usr/local/bin \
  $PKG_SRV/etc/opslens-pulse \
  $PKG_SRV/etc/systemd/system

cp $BUILD/$SERVER_APP $PKG_SRV/usr/local/bin/

cat > $PKG_SRV/DEBIAN/control <<EOF
Package: $SERVER_APP
Version: $VERSION
Section: utils
Priority: optional
Architecture: amd64
Maintainer: Vivek Bangare
Description: OpsLens Pulse Metrics Server
EOF

cat > $PKG_SRV/etc/systemd/system/$SERVER_APP.service <<EOF
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

cat > $PKG_SRV/etc/opslens-pulse/server-config.yaml <<EOF
listen_port: 9898
token: ""
EOF

dpkg-deb --build $PKG_SRV
mv $BUILD/deb/$SERVER_APP.deb $DIST/${SERVER_APP}_${VERSION}_amd64.deb
echo "✅ DEB package for server created"

# ---------------------------
# Checksums
# ---------------------------
cd $DIST
sha256sum * > SHA256SUMS.txt
echo "✅ SHA256SUMS.txt created"

echo "🎉 Release $RAW_VERSION built successfully!"
