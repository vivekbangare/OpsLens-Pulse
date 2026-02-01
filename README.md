# OpsLens-Pulse
OpsLens Pulse - Cross-platform monitoring for hosts, containers, and services on Linux &amp; Windows. Provides real-time metrics via agents. Lightweight, expandable, and visualixed through a built-in dashboard.

### Install result on Linux host

```
sudo dpkg -i opslens-agent_1.0.0_amd64.deb
sudo systemctl enable opslens-agent
sudo systemctl start opslens-agent
```

### 

#### Usage on Windows:

```
.\install.ps1
```


docker build -t deb-builder -f Dockerfile.deb .

docker run --rm \
  -v "$(pwd)":/build \
  -w /build \
  deb-builder \
  ./release.sh 1.0.0

curl -H "Authorization: Bearer mysecrettoken" http://localhost:9898/api/hosts
