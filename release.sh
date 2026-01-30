#!/usr/bin/env bash
set -e

VERSION=$1
if [ -z "$VERSION" ]; then
  echo "Usage: ./release.sh <version>"
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

if [ ! -f "$BUILD/$AGENT_APP" ]; then
  echo "❌ Linux agent build failed!"
  exit 1
fi

echo "🔧 Building Linux server..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
go build -o $BUILD/$SERVER_APP ./server

if [ ! -f "$BUILD/$SERVER_APP" ]; then
  echo "❌ Linux server build failed!"
  exit 1
fi

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

# Copy binary
cp $BUILD/$AGENT_APP $PKG/usr/local/bin/

# Control file
cat > $PKG/DEBIAN/control <<EOF
Package: $AGENT_APP
Version: $VERSION
Section: utils
Priority: optional
Architecture: amd64
Maintainer: Vivek Bangare
Description: opslens Pulse Host Monitoring Agent
EOF

# systemd service
cat > $PKG/etc/systemd/system/$AGENT_APP.service <<EOF
[Unit]
Description=OpsLens Pulse Agent
After=network.target

[Service]
ExecStart=/usr/local/bin/$AGENT_APP --help
EnvironmentFile=/etc/opslens-pulse/agent-config.yaml
Restart=always
RestartSec=5
User=root

[Install]
WantedBy=multi-user.target
EOF

# Default YAML config for Linux
cat > $PKG/etc/opslens-pulse/agent-config.yaml <<EOF
server:
  url: "http://localhost:9898"
  token: "changeme"

agent:
  interval_seconds: 10
  self_upgrade: false
EOF

# Build DEB
dpkg-deb --build $PKG
mv $BUILD/deb/$AGENT_APP.deb $DIST/${AGENT_APP}_${VERSION}_amd64.deb
echo "✅ DEB package for agent created"
echo "➡️  Run '/usr/local/bin/$AGENT_APP --help' to see CLI options."

# ---------------------------
# Build Windows binaries
# ---------------------------
echo "🪟 Building Windows agent..."
(cd agent && GOOS=windows GOARCH=amd64 go build -o ../../$DIST/${AGENT_APP}_${VERSION}_windows_amd64.exe)

WIN_CONFIG_DIR="$DIST/WindowsConfig"
mkdir -p "$WIN_CONFIG_DIR"
cat > "$WIN_CONFIG_DIR/agent-config.yaml" <<EOF
server:
  url: "http://localhost:9898"
  token: "changeme"

agent:
  interval_seconds: 10
  self_upgrade: false
EOF

# Windows installer script with optional custom config path
WIN_INSTALL_SCRIPT="$DIST/install-windows-agent.ps1"
cat > $WIN_INSTALL_SCRIPT <<'EOF'
param(
    [string]$ConfigPath = ""
)

$SourceDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$AgentExe = Join-Path $SourceDir "opslens-pulse-agent_${VERSION}_windows_amd64.exe"
$ConfigSrc = Join-Path $SourceDir "WindowsConfig\agent-config.yaml"

if ($ConfigPath -eq "") {
    $ConfigDestDir = "C:\ProgramData\OpsLens-Pulse"
} else {
    $ConfigDestDir = $ConfigPath
}

if (-Not (Test-Path $ConfigDestDir)) {
    New-Item -ItemType Directory -Path $ConfigDestDir -Force
}

$ConfigDest = Join-Path $ConfigDestDir "agent-config.yaml"

Copy-Item -Path $ConfigSrc -Destination $ConfigDest -Force
Copy-Item -Path $AgentExe -Destination (Join-Path $ConfigDestDir "opslens-pulse-agent.exe") -Force

Write-Host "✅ OpsLens agent installed at $ConfigDestDir"
Write-Host "Config file: $ConfigDest"
Write-Host "Run 'opslens-pulse-agent.exe --help' to see CLI options."
EOF

echo "✅ Windows install script created"

echo "🪟 Building Windows server..."
(cd server && GOOS=windows GOARCH=amd64 go build -o ../../$DIST/${SERVER_APP}_${VERSION}_windows_amd64.exe)
echo "✅ Windows server binary created"

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
token: changeme
EOF

dpkg-deb --build $PKG_SRV
mv $BUILD/deb/$SERVER_APP.deb $DIST/${SERVER_APP}_${VERSION}_amd64.deb
echo "✅ DEB package for server created"


echo "🎉 Release $VERSION built successfully!"
