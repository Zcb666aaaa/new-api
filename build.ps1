Write-Host "Building frontend..." -ForegroundColor Green
Set-Location web
bun install
bun run build
Set-Location ..

Write-Host "Building backend for Linux (amd64)..." -ForegroundColor Green
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"

$VERSION = "dev"
if (Test-Path VERSION) {
    $versionContent = Get-Content VERSION -Raw -ErrorAction SilentlyContinue
    if ($versionContent) {
        $VERSION = $versionContent.Trim()
    }
}

go build -ldflags "-s -w -X 'github.com/QuantumNous/new-api/common.Version=$VERSION'" -o new-api main.go

Write-Host "Build complete! Output: new-api" -ForegroundColor Green
