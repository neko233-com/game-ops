<#
.SYNOPSIS
  Optional, project-level codegraph233 provisioning for flow (心流) analysis.
  NOT covered by this repo's tests. Downloads the matching GitHub release into
  <project>/.codegraph233/bin/ so it never pollutes the global environment.
#>
param(
    [string]$ProjectRoot = (Get-Location).Path,
    [string]$Version = "latest",
    [switch]$SkipIndex
)

$ErrorActionPreference = 'Stop'
$repo = 'neko233-com/codegraph233'
$binDir = Join-Path $ProjectRoot '.codegraph233/bin'
$exe = Join-Path $binDir 'codegraph233.exe'

if (Get-Command codegraph233 -ErrorAction SilentlyContinue) {
    $exe = (Get-Command codegraph233).Source
    Write-Host "Using codegraph233 from PATH: $exe"
} elseif (Test-Path $exe) {
    Write-Host "Using project-local codegraph233: $exe"
} else {
    Write-Host "Downloading $repo ($Version) into $binDir ..." -ForegroundColor Cyan
    New-Item -ItemType Directory -Force -Path $binDir | Out-Null

    $api = if ($Version -eq 'latest') { "https://api.github.com/repos/$repo/releases/latest" } else { "https://api.github.com/repos/$repo/releases/tags/$Version" }
    $release = Invoke-RestMethod -Uri $api -Headers @{ 'User-Agent' = 'game-ops-codegraph-setup' }
    $asset = $release.assets | Where-Object {
        $_.name -match '(?i)win(dows)?.*(x86_64|amd64)| (x86_64|amd64).*win' -and $_.name -match '(?i)\.zip$'
    } | Select-Object -First 1
    if (-not $asset) {
        Write-Warning "No matching Windows release asset found. Install manually:`n  git clone https://github.com/$repo.git`n  cd codegraph233; cargo install --path . --force"
        exit 2
    }
    $zip = Join-Path $binDir $asset.name
    Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $zip -Headers @{ 'User-Agent' = 'game-ops-codegraph-setup' }
    Expand-Archive -Force -Path $zip -DestinationPath $binDir
    Remove-Item $zip -Force
    $candidate = Get-ChildItem -Recurse -Filter 'codegraph233.exe' -Path $binDir | Select-Object -First 1
    if (-not $candidate) { throw "codegraph233.exe not found after extracting $($asset.name)" }
    $exe = $candidate.FullName
    Write-Host "Downloaded: $exe" -ForegroundColor Green
}

$config = Join-Path $ProjectRoot '.codegraph233.json'
if (-not (Test-Path $config)) {
    @'
{
  "version": 1,
  "profile": "full",
  "roots": [{ "key": "client", "path": ".", "role": "frontend" }],
  "languages": ["auto"]
}
'@ | Set-Content -Encoding UTF8 $config
    Write-Host "Wrote $config (engine auto-detection)"
}

& $exe init
& $exe index
& $exe status

Write-Host @'

Next, in your agent, query via codegraph233 MCP:
  cg233_design_graph   - design docs <-> code <-> assets
  cg233_config_tables  - discovered numeric config tables
  cg233_call_graph     - runtime paths / reward loop reachability
  cg233_asset_tree     - content structure supporting flow

Flow analysis is slower than numeric scoring and is NOT tested in this repo.
'@ -ForegroundColor Yellow
