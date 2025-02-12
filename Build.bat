@echo off
echo Building project...

:: Create bin directory if it doesn't exist
if not exist "bin" mkdir bin

:: Build client
echo Building client...
go build -o bin/client.exe ./client
if %ERRORLEVEL% NEQ 0 (
    echo Client build failed!
    pause
    exit /b 1
)

:: Build server
echo Building server...
go build -o bin/server.exe ./server
if %ERRORLEVEL% NEQ 0 (
    echo Server build failed!
    pause
    exit /b 1
)

echo Build successful!
echo Output files:
echo - bin\client.exe
echo - bin\server.exe
pause