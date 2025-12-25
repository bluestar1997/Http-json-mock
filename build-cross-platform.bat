@echo off
echo 开始跨平台编译 HTTP+JSON 工具...

if exist dist rmdir /s /q dist
mkdir dist

echo.
echo 编译 Windows 64位版本...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o dist\http-json-tool-windows-amd64.exe .

echo 编译 Linux 64位版本...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o dist\http-json-tool-linux-amd64 .

echo 编译 macOS 64位版本...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="-s -w" -o dist\http-json-tool-macos-amd64 .

echo 编译 Linux ARM64版本...
set GOOS=linux
set GOARCH=arm64
go build -ldflags="-s -w" -o dist\http-json-tool-linux-arm64 .

echo.
echo 复制必要文件到各平台目录...

REM Windows
mkdir dist\windows
copy dist\http-json-tool-windows-amd64.exe dist\windows\http-json-tool.exe
xcopy json_files dist\windows\json_files\ /E /I
echo @echo off > dist\windows\start.bat
echo echo 启动 HTTP+JSON 工具... >> dist\windows\start.bat
echo echo 请在浏览器中访问: http://localhost:8080 >> dist\windows\start.bat
echo http-json-tool.exe >> dist\windows\start.bat
echo pause >> dist\windows\start.bat

REM Linux
mkdir dist\linux
copy dist\http-json-tool-linux-amd64 dist\linux\http-json-tool
xcopy json_files dist\linux\json_files\ /E /I
echo #!/bin/bash > dist\linux\start.sh
echo echo "启动 HTTP+JSON 工具..." >> dist\linux\start.sh
echo echo "请在浏览器中访问: http://localhost:8080" >> dist\linux\start.sh
echo ./http-json-tool >> dist\linux\start.sh

REM macOS
mkdir dist\macos
copy dist\http-json-tool-macos-amd64 dist\macos\http-json-tool
xcopy json_files dist\macos\json_files\ /E /I
echo #!/bin/bash > dist\macos\start.sh
echo echo "启动 HTTP+JSON 工具..." >> dist\macos\start.sh
echo echo "请在浏览器中访问: http://localhost:8080" >> dist\macos\start.sh
echo ./http-json-tool >> dist\macos\start.sh

echo.
echo 创建说明文件...
echo HTTP+JSON协议收发工具 - 跨平台版本 > dist\README.txt
echo. >> dist\README.txt
echo 支持平台: >> dist\README.txt
echo - Windows 64位: windows目录 >> dist\README.txt
echo - Linux 64位: linux目录 >> dist\README.txt
echo - macOS 64位: macos目录 >> dist\README.txt
echo. >> dist\README.txt
echo 使用方法: >> dist\README.txt
echo Windows: 双击 windows\start.bat >> dist\README.txt
echo Linux/macOS: 运行 chmod +x start.sh && ./start.sh >> dist\README.txt
echo. >> dist\README.txt
echo 访问地址: http://localhost:8080 >> dist\README.txt

echo.
echo ✅ 跨平台编译完成！
echo 📁 文件位置: dist 目录
echo 🖥️  Windows: dist\windows\
echo 🐧 Linux: dist\linux\
echo 🍎 macOS: dist\macos\
echo.
pause