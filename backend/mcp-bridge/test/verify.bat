@echo off
REM MCP Bridge Backend Verification Script (Windows)

echo.
echo MCP Bridge Backend Verification
echo ==================================
echo.

REM Check if we're in the right directory
if not exist "..\..\go.mod" (
    echo [X] Error: Must run from mcp-bridge\test directory
    echo     cd mcp-bridge\test
    echo     verify.bat
    exit /b 1
)

echo [OK] Directory check passed

REM Check if backend can be built
echo.
echo Building backend to verify compilation...
cd ..\..
go build -o temp_test.exe .\cmd\server >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [OK] Backend compiles successfully
    del temp_test.exe >nul 2>&1
) else (
    echo [X] Backend compilation failed
    echo     Run: go build .\cmd\server
    exit /b 1
)

REM Check database schema
echo.
echo Checking database schema...
findstr /C:"mcp_connections" internal\database\database.go >nul
if %ERRORLEVEL% EQU 0 (
    echo [OK] mcp_connections table defined
) else (
    echo [X] mcp_connections table missing
    exit /b 1
)

findstr /C:"mcp_tools" internal\database\database.go >nul
if %ERRORLEVEL% EQU 0 (
    echo [OK] mcp_tools table defined
) else (
    echo [X] mcp_tools table missing
    exit /b 1
)

findstr /C:"mcp_audit_log" internal\database\database.go >nul
if %ERRORLEVEL% EQU 0 (
    echo [OK] mcp_audit_log table defined
) else (
    echo [X] mcp_audit_log table missing
    exit /b 1
)

REM Check MCP models
echo.
echo Checking MCP models...
if exist "internal\models\mcp.go" (
    echo [OK] MCP models file exists
) else (
    echo [X] internal\models\mcp.go missing
    exit /b 1
)

REM Check MCP service
echo.
echo Checking MCP service...
if exist "internal\services\mcp_bridge_service.go" (
    echo [OK] MCP bridge service exists
) else (
    echo [X] internal\services\mcp_bridge_service.go missing
    exit /b 1
)

REM Check MCP handler
echo.
echo Checking MCP handler...
if exist "internal\handlers\mcp_websocket.go" (
    echo [OK] MCP WebSocket handler exists
) else (
    echo [X] internal\handlers\mcp_websocket.go missing
    exit /b 1
)

REM Check tool registry
echo.
echo Checking tool registry extensions...
findstr /C:"RegisterUserTool" internal\tools\registry.go >nul
if %ERRORLEVEL% EQU 0 (
    echo [OK] RegisterUserTool method exists
) else (
    echo [X] RegisterUserTool method missing
    exit /b 1
)

findstr /C:"GetUserTools" internal\tools\registry.go >nul
if %ERRORLEVEL% EQU 0 (
    echo [OK] GetUserTools method exists
) else (
    echo [X] GetUserTools method missing
    exit /b 1
)

REM Check MCP endpoint
echo.
echo Checking MCP endpoint...
findstr /C:"/mcp/connect" cmd\server\main.go >nul
if %ERRORLEVEL% EQU 0 (
    echo [OK] /mcp/connect endpoint registered
) else (
    echo [X] /mcp/connect endpoint missing
    exit /b 1
)

REM Check dependencies
echo.
echo Checking mock client dependencies...
cd mcp-bridge
go list -m github.com/gorilla/websocket >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [OK] gorilla/websocket dependency installed
) else (
    echo [!] gorilla/websocket not installed
    echo     Run: go get github.com/gorilla/websocket
)

REM Final summary
echo.
echo ==================================
echo [OK] All checks passed!
echo.
echo Next steps:
echo   1. Start backend:
echo      cd backend
echo      go run .\cmd\server
echo.
echo   2. Get JWT token from browser (see QUICKSTART.md)
echo.
echo   3. Run mock client:
echo      cd mcp-bridge\test
echo      go run mock_client.go --token YOUR_TOKEN
echo.
echo Documentation:
echo   - Quick test: mcp-bridge\test\QUICKSTART.md
echo   - Full guide: mcp-bridge\test\README.md
echo   - Architecture: mcp-bridge\README.md
echo.
echo Backend is ready for testing!
echo.
