# Verifies the engine and the "test content is never pulled" guarantee.
$ErrorActionPreference = "Stop"
Set-Location (Split-Path -Parent $PSScriptRoot)

$env:CGO_ENABLED = "0"

Write-Host "== go env ==" -ForegroundColor Cyan
go version
go env CGO_ENABLED

Write-Host "== gofmt check ==" -ForegroundColor Cyan
$unformatted = gofmt -l (Get-ChildItem -Recurse -Filter *.go | ForEach-Object FullName)
if ($unformatted) { throw "Unformatted files:`n$unformatted" }
Write-Host "all Go files formatted"

Write-Host "== go vet ==" -ForegroundColor Cyan
go vet ./...
if ($LASTEXITCODE -ne 0) { throw "go vet failed" }

Write-Host "== go build (no cgo) ==" -ForegroundColor Cyan
go build ./...
if ($LASTEXITCODE -ne 0) { throw "go build failed" }

Write-Host "== go test ==" -ForegroundColor Cyan
go test ./...
if ($LASTEXITCODE -ne 0) { throw "go test failed" }

Write-Host "== gitignore guarantee ==" -ForegroundColor Cyan
git check-ignore -v --no-index tests/num_test.go
git status --ignored --short
