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
