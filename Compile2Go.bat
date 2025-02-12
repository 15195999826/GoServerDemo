@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

REM 设置 schema 目录路径和生成目录路径
set SCHEMA_DIR=.\schema
set OUTPUT_DIR=.\fb

REM 检查并创建输出目录
if not exist %OUTPUT_DIR% mkdir %OUTPUT_DIR%

REM 删除输出目录中的所有 .go 文件
echo 正在清理旧文件...
del /Q %OUTPUT_DIR%\*.go 2>nul
if errorlevel 1 (
    echo 清理文件时出错或目录为空
) else (
    echo 成功清理旧文件
)

REM 遍历所有 .fbs 文件并使用 flatc 编译
echo 开始编译 .fbs 文件...
for /r %SCHEMA_DIR% %%i in (*.fbs) do (
    REM 检查是否执行报错， echo不同结果
    flatc --go %%i
    if errorlevel 1 (
        echo 编译失败: %%i
    ) else (
        echo 编译成功: %%i
    )
)

echo 完成所有文件的编译。
pause