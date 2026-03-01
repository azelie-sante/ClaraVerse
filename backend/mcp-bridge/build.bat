@echo off
REM Cross-platform build script for Clara Companion

echo Building Clara Companion...
echo.

REM Windows
echo [1/3] Building for Windows (amd64)...
set GOOS=windows
set GOARCH=amd64
go build -o bin/clara_companion-windows-amd64.exe ./cmd/clara_companion
if %ERRORLEVEL% NEQ 0 (
    echo Build failed for Windows
    exit /b 1
)
echo ✓ Windows build complete

REM Linux
echo [2/3] Building for Linux (amd64)...
set GOOS=linux
set GOARCH=amd64
go build -o bin/clara_companion-linux-amd64 ./cmd/clara_companion
if %ERRORLEVEL% NEQ 0 (
    echo Build failed for Linux
    exit /b 1
)
echo ✓ Linux build complete

REM macOS
echo [3/3] Building for macOS (amd64)...
set GOOS=darwin
set GOARCH=amd64
go build -o bin/clara_companion-darwin-amd64 ./cmd/clara_companion
if %ERRORLEVEL% NEQ 0 (
    echo Build failed for macOS
    exit /b 1
)
echo ✓ macOS build complete

echo.
echo ========================================
echo ✓ All builds completed successfully!
echo ========================================
echo.
echo Binaries created in bin/:
dir /B bin\clara_companion-*
echo.
echo To run:
echo   Windows: bin\clara_companion-windows-amd64.exe
echo   Linux:   bin/clara_companion-linux-amd64
echo   macOS:   bin/clara_companion-darwin-amd64
echo.
