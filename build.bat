@echo off
echo 开始编译 HTTP+JSON 工具...
go build -ldflags "-s -w" -o http-json-tool.exe .

if %ERRORLEVEL% neq 0 (
    echo 编译失败！
    pause
    exit /b 1
)

echo 创建发布目录...
if exist release rmdir /s /q release
mkdir release

echo 复制文件...
copy http-json-tool.exe release\
xcopy json_files release\json_files\ /E /I
if exist config.json copy config.json release\

echo 创建启动脚本...
echo @echo off > release\start.bat
echo echo 启动 HTTP+JSON 工具... >> release\start.bat
echo echo 请在浏览器中访问: http://localhost:8080 >> release\start.bat
echo http-json-tool.exe >> release\start.bat
echo pause >> release\start.bat

echo 创建说明文件...
echo HTTP+JSON协议收发工具 > release\README.txt
echo. >> release\README.txt
echo 使用方法: >> release\README.txt
echo 1. 双击 start.bat 启动程序 >> release\README.txt
echo 2. 在浏览器中访问 http://localhost:8080 >> release\README.txt
echo 3. 按 Ctrl+C 或关闭窗口停止程序 >> release\README.txt
echo. >> release\README.txt
echo 新增功能: >> release\README.txt
echo - 配置文件自动保存: 程序会自动保存服务器配置到 config.json 文件 >> release\README.txt
echo - 启动时自动加载配置: 程序启动时会读取 config.json 文件恢复之前的设置 >> release\README.txt
echo. >> release\README.txt
echo 功能说明: >> release\README.txt
echo - 接收部分: 配置HTTP服务器，设置接口响应 >> release\README.txt
echo - 发送部分: 发送HTTP请求到其他服务 >> release\README.txt
echo - JSON编辑: 选择响应文件后可在线编辑内容 >> release\README.txt
echo - 配置持久化: IP、端口、接口路径等配置会自动保存到本地文件 >> release\README.txt

echo.
echo ✅ 打包完成！
echo 📁 文件位置: release 目录
echo 🚀 运行方式: 双击 release\start.bat
echo 🌐 访问地址: http://localhost:8080
echo.
pause