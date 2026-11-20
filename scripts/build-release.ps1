<#
.SYNOPSIS
Build production release of the single `gameops` binary for multiple platforms.
.DESCRIPTION
Produces dist\<version>\<os>-<arch>\gameops[.exe] + SHA256SUMS + version.json.
.PARAMETER Version
Semantic version tag, default v1.0.0.
.PARAMETER Install
Copy the native Windows amd64 binary to an install dir (default $HOME\bin).
#>
param(
  [string]$Version = "v1.0.0",
  [switch]$Install,
  [string]$InstallDir = "$HOME\bin"
)

$ErrorActionPreference = "Stop"
$RepoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $RepoRoot

$commit = ""
try { $commit = (& git rev-parse --short HEAD 2>$null).Trim() } catch {}
if (-not $commit) { $commit = "none" }
$date = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$ldflags = "-X main.version=$Version -X main.commit=$commit -X main.date=$date"
$out = "dist/$Version"

Write-Host "Building gameops $Version ($commit) ..."
if (Test-Path $out) { Remove-Item -Recurse -Force $out }
New-Item -ItemType Directory -Force -Path $out | Out-Null

$targets = @(
  @{ os="windows"; arch="amd64" },
  @{ os="windows"; arch="arm64" },
  @{ os="linux";   arch="amd64" },
  @{ os="linux";   arch="arm64" },
  @{ os="darwin";  arch="amd64" },
  @{ os="darwin";  arch="arm64" }
)

foreach ($t in $targets) {
  $ext = if ($t.os -eq "windows") { ".exe" } else { "" }
  $bin = "gameops-$($t.os)-$($t.arch)$ext"
  $dest = Join-Path $out $bin
  Write-Host "  -> $bin"
  $env:CGO_ENABLED = "0"; $env:GOOS = $t.os; $env:GOARCH = $t.arch
  & go build -trimpath -ldflags $ldflags -o $dest ./cmd/gameops
  if ($LASTEXITCODE -ne 0) { throw "build failed for $($t.os)/$($t.arch)" }
}
Remove-Item Env:GOOS, Env:GOARCH -ErrorAction SilentlyContinue
$env:CGO_ENABLED = "0"

$utf8NoBom = New-Object System.Text.UTF8Encoding($false)
# SHA256SUMS over the binaries, flat names matching GitHub Release assets
$sums = Get-ChildItem -File $out -Filter "gameops-*" | Sort-Object Name | ForEach-Object {
  $hash = (Get-FileHash -Algorithm SHA256 $_.FullName).Hash.ToLower()
  "$hash  $($_.Name)"
}
[System.IO.File]::WriteAllText((Join-Path $out "SHA256SUMS"), (($sums -join "`n") + "`n"), $utf8NoBom)

$goVer = (& go version).Split()[2]
$versionInfo = [ordered]@{
  version = $Version
  commit  = $commit
  date    = $date
  go      = $goVer
  targets = @("windows-amd64","windows-arm64","linux-amd64","linux-arm64","darwin-amd64","darwin-arm64")
}
[System.IO.File]::WriteAllText((Join-Path $out "version.json"), ($versionInfo | ConvertTo-Json -Depth 3), $utf8NoBom)

Write-Host "Done: $out"
Write-Host "Verify: Get-Content $out/SHA256SUMS"

if ($Install) {
  New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
  $src = Join-Path $out "gameops-windows-amd64.exe"
  Copy-Item -Force $src (Join-Path $InstallDir "gameops.exe")
  Write-Host "Installed: $InstallDir/gameops.exe"
  Write-Host "Add $InstallDir to PATH to run 'gameops' anywhere."
}
