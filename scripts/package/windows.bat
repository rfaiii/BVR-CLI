@echo off
setlocal

set ROOT=%~dp0..
set VERSION=%~1
if "%VERSION%"=="" set VERSION=dev

set BUILD_DIR=%ROOT%\dist\windows
set EXE=%BUILD_DIR%\bvr-cli.exe
set ZIP=%BUILD_DIR%\bvr-cli_%VERSION%_windows_amd64.zip

if not exist "%BUILD_DIR%" mkdir "%BUILD_DIR%"

set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
set GOEXPERIMENT=greenteagc

go build -trimpath -ldflags="-s -w -X github.com/richavery/bvr-cli/internal/version.Version=%VERSION%" -o "%EXE%" "%ROOT%"

copy /Y "%ROOT%\README.md" "%BUILD_DIR%\README.txt" >nul
copy /Y "%ROOT%\LICENSE" "%BUILD_DIR%\LICENSE.txt" >nul 2>nul
copy /Y "%ROOT%\LICENSE.md" "%BUILD_DIR%\LICENSE.txt" >nul 2>nul

powershell -Command "Compress-Archive -Path '%EXE%','%BUILD_DIR%\README.txt','%BUILD_DIR%\LICENSE.txt' -DestinationPath '%ZIP%' -Force"

echo Packaged Windows zip at: %ZIP%
endlocal
