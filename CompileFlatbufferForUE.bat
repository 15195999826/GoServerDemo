@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

REM Set directory paths (use quotes to handle spaces)
set "SCHEMA_DIR=%~dp0schema"
set "UE_PROJECT_DIR=C:\UEProjects\DESKTK"
set "SOURCE_DIR=%UE_PROJECT_DIR%\Source\DESKTK"
set "OUTPUT_DIR=%SOURCE_DIR%\Generated"

REM Check and create output directory
if not exist "%OUTPUT_DIR%" (
    mkdir "%OUTPUT_DIR%"
    if !ERRORLEVEL! NEQ 0 (
        echo Error: Failed to create output directory
        pause
        exit /b 1
    )
)

REM Clean old files
echo Cleaning old files...
del /Q "%OUTPUT_DIR%\*.h" 2>nul

REM Compile .fbs files to C++
echo Compiling .fbs files...
for /r "%SCHEMA_DIR%" %%i in (*.fbs) do (
    echo Processing: "%%i"
    flatc --cpp --cpp-std c++17 --scoped-enums --gen-object-api --gen-compare ^
          --cpp-ptr-type "std::unique_ptr" ^
          --cpp-static-reflection --filename-suffix "" ^
          -o "%OUTPUT_DIR%" "%%i"
    if !ERRORLEVEL! NEQ 0 (
        echo Error: Failed to compile "%%i"
        pause
        exit /b 1
    ) else (
        echo Successfully compiled "%%i"
    )
)

echo.
echo Generation completed successfully.
echo Files have been output to: %OUTPUT_DIR%
echo Please include the generated headers in your UE project.
pause