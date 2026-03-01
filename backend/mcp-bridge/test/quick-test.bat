@echo off
REM Quick MCP Bridge Test Script for Windows
REM This script helps you test the MCP Bridge end-to-end

echo.
echo ========================================
echo   MCP Bridge Quick Test (Windows)
echo ========================================
echo.

REM Step 1: Build the client
echo [1/5] Building MCP client...
cd ..
if exist "mcp-client.exe" del mcp-client.exe
go build -o mcp-client.exe .\cmd\mcp-client
if %ERRORLEVEL% NEQ 0 (
    echo [X] Build failed!
    pause
    exit /b 1
)
echo [OK] Client built successfully
echo.

REM Step 2: Check version
echo [2/5] Checking client version...
.\mcp-client.exe --version
if %ERRORLEVEL% NEQ 0 (
    echo [X] Client won't run!
    pause
    exit /b 1
)
echo.

REM Step 3: Instructions for backend
echo [3/5] Backend Check
echo.
echo Is your backend running on port 3001?
echo If not, open a NEW terminal and run:
echo    cd backend
echo    go run .\cmd\server
echo.
pause

REM Step 4: Get JWT Token
echo.
echo [4/5] Getting JWT Token
echo.
echo Open your browser and follow these steps:
echo   1. Go to http://localhost:5173
echo   2. Login to ClaraVerse
echo   3. Press F12 (open DevTools)
echo   4. Go to Console tab
echo   5. Run: localStorage.getItem('supabase.auth.token')
echo   6. Copy the token (starts with ey...)
echo.
echo When ready, run the login command:
echo    .\mcp-client.exe login
echo.
pause

REM Step 5: Instructions for running
echo.
echo [5/5] Testing Instructions
echo.
echo Now you can test the client! Run these commands:
echo.
echo 1. Login (if you haven't already):
echo    .\mcp-client.exe login
echo.
echo 2. Check status:
echo    .\mcp-client.exe status
echo.
echo 3. Start the client:
echo    .\mcp-client.exe start
echo.
echo 4. Open browser and chat:
echo    - Go to http://localhost:5173
echo    - Start a NEW conversation
echo    - Ask: "What tools do you have?"
echo    - You should see any MCP tools you've added
echo.
echo ========================================
echo   Setup Complete!
echo ========================================
echo.
echo Client binary: %cd%\mcp-client.exe
echo Config will be: %USERPROFILE%\.claraverse\mcp-config.yaml
echo.
echo For mock testing (without real MCP servers):
echo    cd test
echo    go run mock_client.go --token YOUR_TOKEN
echo.
pause
