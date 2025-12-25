# HTTP+JSON协议收发工具

这是一个基于Go语言开发的简单HTTP+JSON协议收发程序，提供了直观的Web界面和强大的功能。

## 功能特性

1. **简洁的Web用户界面** - 操作方便，界面友好
2. **灵活的服务器配置** - 可指定监听IP和端口（如：http://192.168.206.187:29800）
3. **多接口支持** - 同时支持最多5个自定义接口的请求接收和内容显示
4. **JSON响应文件** - 可为每个接口选择对应的JSON文件作为响应数据
5. **HTTP请求发送** - 支持向指定地址发送JSON数据
6. **在线JSON编辑** - 支持在线编辑和格式化JSON数据
7. **实时请求监控** - 通过WebSocket实时显示接收到的请求
8. **请求日志记录** - 详细记录所有接收到的HTTP请求

## 项目结构

```
http-json-tool/
├── main.go              # 主程序文件
├── ui.go               # 用户界面生成文件
├── go.mod              # Go模块依赖文件
├── templates/          # HTML模板目录
│   └── index.html      # 主页面模板
├── static/             # 静态资源目录
│   ├── style.css       # 样式文件
│   └── app.js          # JavaScript应用文件
├── json_files/         # JSON响应文件目录
│   ├── success_response.json
│   ├── error_response.json
│   ├── audit_task_result.json
│   ├── user_list.json
│   └── config_data.json
└── README.md           # 使用说明文件
```

## 快速开始

### 1. 环境要求

- Go 1.21 或更高版本
- 网络连接（用于下载依赖）

### 2. 安装和运行

```bash
# 进入项目目录
cd http-json-tool

# 下载依赖（如果还未执行）
go mod tidy

# 运行程序
go run .
```

### 3. 访问界面

程序启动后，在浏览器中访问：
```
http://localhost:8080
```

## 使用说明

### 服务器配置

1. **设置监听地址**：在界面上方输入要监听的IP地址和端口号
2. **启动服务器**：点击"启动服务器"按钮
3. **停止服务器**：点击"停止服务器"按钮

### 接口配置

1. **启用/禁用接口**：勾选复选框来启用或禁用特定接口
2. **设置响应文件**：为每个接口选择对应的JSON响应文件
3. **保存配置**：点击"保存配置"按钮保存设置

默认包含的接口：
- `/api/audioTask/getAuditTaskResult`
- `/api/test1`
- `/api/test2`
- `/api/test3`
- `/api/test4`

### 发送HTTP请求

1. **输入目标URL**：在请求URL字段输入完整的HTTP地址
2. **选择请求方法**：支持GET、POST、PUT、DELETE
3. **设置请求头**：以JSON格式输入自定义请求头
4. **输入请求数据**：以JSON格式输入要发送的数据
5. **发送请求**：点击"发送请求"按钮
6. **查看响应**：响应结果将显示在下方区域

### 请求日志

- 所有接收到的HTTP请求都会实时显示在日志区域
- 点击"详情"按钮可查看完整的请求头和请求体
- 支持清空日志和刷新日志功能

### JSON文件管理

在`json_files/`目录下放置您的JSON响应文件，程序会自动扫描并在接口配置中提供选择。

## 示例用法

### 1. 作为Mock服务器

1. 启动程序并配置监听地址为`192.168.1.100:29800`
2. 启用`/api/audioTask/getAuditTaskResult`接口
3. 为该接口选择`audit_task_result.json`作为响应文件
4. 其他应用可以通过`http://192.168.1.100:29800/api/audioTask/getAuditTaskResult`访问并获得JSON响应

### 2. 作为HTTP客户端

1. 在发送请求区域输入目标URL：`http://api.example.com/users`
2. 选择POST方法
3. 设置请求头：`{"Content-Type": "application/json", "Authorization": "Bearer token123"}`
4. 输入请求数据：`{"name": "张三", "email": "zhangsan@example.com"}`
5. 点击发送请求查看响应结果

## 技术栈

- **后端**：Go + Gin框架
- **前端**：原生HTML/CSS/JavaScript
- **实时通信**：WebSocket (gorilla/websocket)
- **HTTP客户端**：Go标准库net/http

## 注意事项

- 程序启动时会自动创建必要的目录和示例文件
- 请求日志最多保存100条记录，超出后会自动清理旧记录
- WebSocket连接断开时会自动重连
- JSON文件格式必须正确，否则会返回默认响应

## 许可证

本项目采用MIT许可证。