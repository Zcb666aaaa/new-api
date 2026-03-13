@echo off
echo Building frontend...


echo Building backend for Linux (amd64)...
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64

set VERSION=dev
if exist VERSION (
    for /f "delims=" %%i in (VERSION) do set VERSION=%%i
)
go build -ldflags "-s -w -X 'github.com/QuantumNous/new-api/common.Version=%VERSION%'" -o new-api main.go

echo Build complete! Output: new-api
