param(
    [ValidateSet('project', 'global', 'all')]
    [string]$Scope = 'project'
)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$source = Join-Path $repoRoot 'skills/game-long-term-operation'
if (-not (Test-Path (Join-Path $source 'SKILL.md'))) {
    throw "Canonical skill not found at $source"
}

function Find-ProjectRoot {
    $dir = (Get-Location).Path
    while ($true) {
        foreach ($marker in @('.git', 'go.mod', 'package.json', '.opencode', '.claude', '.agents', '.dsh')) {
            if (Test-Path (Join-Path $dir $marker)) { return $dir }
        }
        $parent = Split-Path -Parent $dir
        if ([string]::IsNullOrEmpty($parent) -or $parent -eq $dir) { return $null }
        $dir = $parent
    }
}

$projectRoot = Find-ProjectRoot
$homeDir = $env:USERPROFILE
$runtimes = @(
    @{ Name = 'opencode'; Project = '.opencode/skills'; Global = '.config/opencode/skills' },
    @{ Name = 'claude';   Project = '.claude/skills';   Global = '.claude/skills' },
    @{ Name = 'codex';    Project = '.agents/skills';   Global = '.agents/skills' },
    @{ Name = 'dsh';      Project = '.dsh/skills';      Global = '.dsh/skills' }
)

if ($Scope -eq 'project' -and -not $projectRoot) {
    Write-Host "No project detected; falling back to global install." -ForegroundColor Yellow
    $Scope = 'global'
}

$installed = @()
foreach ($rt in $runtimes) {
    $targets = @()
    if ($Scope -in @('project', 'all') -and $projectRoot) {
        $targets += Join-Path $projectRoot ($rt.Project -replace '/', [IO.Path]::DirectorySeparatorChar)
    }
    if ($Scope -in @('global', 'all')) {
        $globalEnv = ($rt.Name.ToUpper() + '_SKILLS_DIR')
        $override = [Environment]::GetEnvironmentVariable($globalEnv)
        if ($rt.Name -eq 'dsh' -and -not $override) {
            $override = [Environment]::GetEnvironmentVariable('SKILLS_DIR')
        }
        $globalBase = if ($override) { $override } else { Join-Path $homeDir ($rt.Global -replace '/', [IO.Path]::DirectorySeparatorChar) }
        $targets += Join-Path $globalBase 'game-long-term-operation'
    }

    foreach ($parent in ($targets | Select-Object -Unique)) {
        $target = Join-Path $parent 'game-long-term-operation'
        if (Test-Path $target) { Remove-Item -Recurse -Force $target }
        New-Item -ItemType Directory -Force -Path $parent | Out-Null
        Copy-Item -Recurse -Force -Path $source -Destination $target
        if (-not (Test-Path (Join-Path $target 'SKILL.md'))) { throw "Install verification failed: $target" }
        $installed += "$($rt.Name): $target"
    }
}

Write-Host "Installed game-long-term-operation ($Scope):" -ForegroundColor Green
$installed | ForEach-Object { Write-Host "  $_" }
if (-not $installed) { Write-Host "  (nothing to install)" }
