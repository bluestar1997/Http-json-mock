package main

import (
	"os"
	"path/filepath"
)

func createHTMLTemplate() {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>HTTP+JSON协议收发工具</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <!-- TAB导航 -->
        <div class="tab-nav">
            <button class="tab-btn active" data-tab="receive">接收部分</button>
            <button class="tab-btn" data-tab="send">发送部分</button>
        </div>

        <div class="main-content">
            <!-- 接收部分TAB -->
            <div class="tab-content active" id="receive-tab">
                <!-- 项目配置区域 -->
                <section class="section compact">
                    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px;">
                        <h2 style="margin: 0;">项目配置</h2>
                        <div style="display: flex; gap: 10px; align-items: center;">
                            <label style="margin: 0;">项目:</label>
                            <select id="project-select" onchange="switchProject()" style="padding: 5px 10px; border: 1px solid #ddd; border-radius: 4px; background: white;">
                                <!-- 动态填充项目列表 -->
                            </select>
                            <button onclick="createNewProject()" class="btn btn-info" style="padding: 5px 15px; white-space: nowrap;">+ 新建项目</button>
                            <button onclick="saveProjectConfig()" class="btn btn-primary" style="padding: 5px 15px; white-space: nowrap;">保存项目配置</button>
                        </div>
                    </div>
                    <div class="config-form">
                        <div class="form-row">
                            <div class="form-group">
                                <label>监听地址:</label>
                                <input type="text" id="server-ip" value="192.168.1.100" placeholder="IP地址">
                            </div>
                            <div class="form-group">
                                <label>端口:</label>
                                <input type="text" id="server-port" value="29800" placeholder="端口">
                            </div>
                            <div class="form-actions">
                                <button id="start-server" class="btn btn-primary">启动</button>
                                <button id="stop-server" class="btn btn-secondary" disabled>停止</button>
                            </div>
                            <div class="status-inline">
                                <span id="server-status" class="status-stopped">已停止</span>
                                <span id="server-url"></span>
                                <span id="error-message" class="error-message" style="display: none;"></span>
                            </div>
                        </div>
                    </div>
                </section>

                <!-- 接口配置区域 -->
                <section class="section compact">
                    <h2>接口配置 (最多5个)</h2>
                    <div id="endpoints-config" class="endpoints-grid-2cols">
                        <!-- 动态生成接口配置 -->
                    </div>

                    <!-- JSON文件编辑区域 -->
                    <div id="json-editor" class="json-editor" style="display: none;">
                        <div class="json-editor-header">
                            <h3 id="json-file-name">编辑JSON文件</h3>
                            <div class="json-editor-actions">
                                <button id="save-json" class="btn btn-primary">保存</button>
                                <button id="close-json-editor" class="btn btn-secondary">关闭</button>
                            </div>
                        </div>
                        <textarea id="json-content" class="json-textarea" placeholder="JSON内容..."></textarea>
                    </div>
                </section>

                <!-- 接收日志区域 -->
                <section class="section">
                    <h2>接收日志 (最近收到的请求)</h2>
                    <div class="logs-container">
                        <div class="logs-actions">
                            <button id="clear-logs" class="btn btn-secondary">清空</button>
                            <button id="refresh-logs" class="btn btn-info">刷新</button>
                        </div>
                        <div id="request-logs" class="logs-grid"></div>
                    </div>
                </section>
            </div>

            <!-- 发送部分TAB -->
            <div class="tab-content" id="send-tab">
                <!-- 发送请求区域 -->
                <section class="section compact">
                    <h2>发送HTTP请求</h2>
                    <div class="request-form">
                        <div class="form-row">
                            <div class="form-group flex-grow">
                                <label>URL:</label>
                                <input type="text" id="request-url" placeholder="http://example.com/api/test">
                            </div>
                            <div class="form-group">
                                <label>方法:</label>
                                <select id="request-method">
                                    <option value="GET">GET</option>
                                    <option value="POST">POST</option>
                                    <option value="PUT">PUT</option>
                                    <option value="DELETE">DELETE</option>
                                </select>
                            </div>
                            <div class="form-group">
                                <label>请求头:</label>
                                <input type="text" id="request-headers" placeholder='{"Content-Type": "application/json"}'>
                            </div>
                        </div>
                        <div class="form-group">
                            <label>请求数据 (JSON):</label>
                            <textarea id="request-data" placeholder='{"key": "value"}' rows="6"></textarea>
                        </div>
                        <div class="form-actions">
                            <button id="send-request" class="btn btn-primary">发送</button>
                            <button id="format-json" class="btn btn-info">格式化</button>
                        </div>
                    </div>

                    <!-- 响应结果 -->
                    <div class="response-section">
                        <h3>响应结果:</h3>
                        <pre id="response-result"></pre>
                    </div>
                </section>
            </div>
        </div>
    </div>

    <script src="/static/app.js"></script>
</body>
</html>`

	os.WriteFile("templates/index.html", []byte(html), 0644)
}

func createCSSFile() {
	css := `* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
    background-color: #f5f5f5;
    color: #333;
    line-height: 1.6;
}

.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 20px;
}

header {
    text-align: center;
    margin-bottom: 30px;
}

header h1 {
    color: #2c3e50;
    font-size: 2.5em;
    margin-bottom: 10px;
}

.main-content {
    display: grid;
    gap: 30px;
}

.section {
    background: white;
    border-radius: 8px;
    padding: 25px;
    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
}

.section h2 {
    color: #2c3e50;
    border-bottom: 3px solid #3498db;
    padding-bottom: 10px;
    margin-bottom: 20px;
    font-size: 1.5em;
}

.section h3 {
    color: #34495e;
    margin: 15px 0 10px;
}

.config-form {
    display: grid;
    gap: 15px;
}

.compact {
    padding: 15px;
}

.form-row {
    display: flex;
    gap: 15px;
    align-items: end;
    flex-wrap: wrap;
}

.form-row .form-group {
    flex: 0 0 auto;
}

.form-row .flex-grow {
    flex: 1 1 auto;
    min-width: 200px;
}

.endpoints-grid {
    display: grid;
    gap: 10px;
}

.endpoints-grid-2cols {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
}

/* TAB导航样式 */
.tab-nav {
    display: flex;
    background: #f8f9fa;
    border-bottom: 2px solid #dee2e6;
    margin-bottom: 20px;
    border-radius: 8px 8px 0 0;
    overflow: hidden;
}

.tab-btn {
    flex: 1;
    padding: 15px 20px;
    border: none;
    background: #f8f9fa;
    color: #666;
    font-size: 16px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.3s;
    border-bottom: 3px solid transparent;
}

.tab-btn:hover {
    background: #e9ecef;
    color: #495057;
}

.tab-btn.active {
    background: white;
    color: #2c3e50;
    border-bottom-color: #3498db;
}

/* TAB内容样式 */
.tab-content {
    display: none;
}

.tab-content.active {
    display: block;
    animation: fadeIn 0.3s ease-in;
}

/* JSON编辑器样式 */
.json-editor {
    margin-top: 15px;
    border: 1px solid #ddd;
    border-radius: 5px;
    background: white;
    box-shadow: 0 2px 5px rgba(0,0,0,0.1);
}

.json-editor-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px 15px;
    background: #f8f9fa;
    border-bottom: 1px solid #ddd;
    border-radius: 5px 5px 0 0;
}

.json-editor-header h3 {
    margin: 0;
    color: #2c3e50;
    font-size: 14px;
}

.json-editor-actions {
    display: flex;
    gap: 8px;
}

.json-editor-actions .btn {
    padding: 5px 12px;
    font-size: 12px;
}

.json-textarea {
    width: 100%;
    height: 200px;
    padding: 15px;
    border: none;
    border-radius: 0 0 5px 5px;
    font-family: 'Courier New', monospace;
    font-size: 12px;
    resize: vertical;
    min-height: 150px;
}

.logs-grid {
    display: grid;
    gap: 8px;
}

.form-group {
    display: flex;
    flex-direction: column;
    gap: 5px;
}

.form-group label {
    font-weight: 600;
    color: #2c3e50;
}

.form-group input,
.form-group select,
.form-group textarea {
    padding: 10px;
    border: 2px solid #ddd;
    border-radius: 5px;
    font-size: 14px;
    transition: border-color 0.3s;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
    outline: none;
    border-color: #3498db;
}

.form-group textarea {
    resize: vertical;
    min-height: 100px;
    font-family: 'Courier New', monospace;
}

.form-actions {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
}

.btn {
    padding: 10px 20px;
    border: none;
    border-radius: 5px;
    cursor: pointer;
    font-size: 14px;
    font-weight: 600;
    transition: all 0.3s;
    text-transform: uppercase;
}

.btn:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(0,0,0,0.2);
}

.btn-primary {
    background-color: #3498db;
    color: white;
}

.btn-primary:hover {
    background-color: #2980b9;
}

.btn-secondary {
    background-color: #95a5a6;
    color: white;
}

.btn-secondary:hover {
    background-color: #7f8c8d;
}

.btn-info {
    background-color: #17a2b8;
    color: white;
}

.btn-info:hover {
    background-color: #138496;
}

.btn-danger {
    background-color: #e74c3c;
    color: white;
}

.btn-danger:hover {
    background-color: #c0392b;
}

.btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
    transform: none !important;
    box-shadow: none !important;
}

.status {
    display: flex;
    align-items: center;
    gap: 15px;
    margin-top: 15px;
}

.status-inline {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 10px;
    margin-left: 15px;
    font-size: 13px;
}

.status-running {
    color: #27ae60;
    font-weight: bold;
}

.status-stopped {
    color: #e74c3c;
    font-weight: bold;
}

.endpoint-item {
    border: 1px solid #ddd;
    border-radius: 5px;
    padding: 8px;
    background: #fafafa;
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
}

.endpoint-item.active {
    border-color: #3498db;
    background: #f8f9fa;
}

.endpoint-checkbox {
    flex: 0 0 auto;
}

.endpoint-path {
    flex: 1 1 auto;
    min-width: 120px;
}

.endpoint-path input {
    width: 100%;
    padding: 4px 6px;
    border: 1px solid #ddd;
    border-radius: 3px;
    font-size: 12px;
}

.endpoint-file {
    flex: 0 0 auto;
    min-width: 100px;
}

.endpoint-file select {
    flex: 1;
    padding: 4px;
    border: 1px solid #ddd;
    border-radius: 3px;
    font-size: 11px;
}

.endpoint-file {
    display: flex;
    gap: 4px;
    align-items: center;
}

.btn-edit {
    padding: 3px 6px;
    border: 1px solid #3498db;
    background: #3498db;
    color: white;
    border-radius: 3px;
    cursor: pointer;
    font-size: 10px;
    text-decoration: none;
    white-space: nowrap;
}

.btn-edit:hover {
    background: #2980b9;
    border-color: #2980b9;
}

.error-message {
    color: #e74c3c;
    font-weight: bold;
    margin-left: 10px;
}

.file-select {
    display: flex;
    gap: 10px;
    align-items: center;
}

.request-form {
    display: grid;
    gap: 15px;
}

.response-section {
    margin-top: 20px;
}

#response-result {
    background: #2c3e50;
    color: #ecf0f1;
    padding: 15px;
    border-radius: 5px;
    max-height: 400px;
    overflow-y: auto;
    font-family: 'Courier New', monospace;
    font-size: 12px;
    white-space: pre-wrap;
}

.logs-container {
    max-height: 500px;
    overflow-y: auto;
}

.logs-actions {
    margin-bottom: 15px;
    display: flex;
    gap: 10px;
}

#request-logs {
    display: grid;
    gap: 10px;
}

.log-item {
    border: 1px solid #ddd;
    border-radius: 5px;
    padding: 10px;
    background: #f9f9f9;
    font-size: 13px;
}

.log-item:hover {
    background: #f0f0f0;
}

.log-header {
    display: grid;
    grid-template-columns: auto auto 1fr auto auto;
    gap: 10px;
    align-items: center;
    margin-bottom: 5px;
    font-weight: bold;
}

.log-content {
    background: #fff;
    padding: 8px;
    border-radius: 3px;
    margin-top: 5px;
    font-family: 'Courier New', monospace;
    font-size: 11px;
    max-height: 120px;
    overflow-y: auto;
    word-break: break-all;
}

.log-method {
    padding: 3px 8px;
    border-radius: 3px;
    color: white;
    font-size: 12px;
    font-weight: bold;
}

.log-method.GET { background-color: #28a745; }
.log-method.POST { background-color: #ffc107; color: #000; }
.log-method.PUT { background-color: #17a2b8; }
.log-method.DELETE { background-color: #dc3545; }

.log-details {
    font-family: 'Courier New', monospace;
    font-size: 12px;
    background: #f8f9fa;
    padding: 10px;
    border-radius: 3px;
    margin-top: 10px;
    max-height: 200px;
    overflow-y: auto;
}

.toggle-btn {
    background: none;
    border: 1px solid #3498db;
    color: #3498db;
    padding: 5px 10px;
    border-radius: 3px;
    cursor: pointer;
    font-size: 12px;
}

.toggle-btn:hover {
    background-color: #3498db;
    color: white;
}

/* 响应式设计 */
@media (max-width: 768px) {
    .container {
        padding: 10px;
    }

    .form-actions {
        flex-direction: column;
    }

    .btn {
        width: 100%;
    }

    .status {
        flex-direction: column;
        align-items: flex-start;
        gap: 5px;
    }

    .endpoint-controls {
        grid-template-columns: 1fr;
        gap: 15px;
    }

    .log-header {
        flex-direction: column;
        align-items: flex-start;
        gap: 5px;
    }
}

/* 滚动条样式 */
::-webkit-scrollbar {
    width: 8px;
}

::-webkit-scrollbar-track {
    background: #f1f1f1;
    border-radius: 10px;
}

::-webkit-scrollbar-thumb {
    background: #888;
    border-radius: 10px;
}

::-webkit-scrollbar-thumb:hover {
    background: #555;
}

/* 动画 */
@keyframes fadeIn {
    from { opacity: 0; transform: translateY(20px); }
    to { opacity: 1; transform: translateY(0); }
}

.section {
    animation: fadeIn 0.5s ease-out;
}

/* JSON语法高亮 */
.json-string { color: #d14; }
.json-number { color: #099; }
.json-boolean { color: #0969da; }
.json-null { color: #6f42c1; }
.json-key { color: #e36209; }`

	os.WriteFile("static/style.css", []byte(css), 0644)
}

func createJSFile() {
	js := `class HTTPJSONTool {
    constructor() {
        this.ws = null;
        this.init();
        this.connectWebSocket();
    }

    init() {
        this.bindEvents();
        this.loadProjects();
        this.loadStatus();
        this.loadJSONFiles();
        this.initTabs();
    }

    connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = ` + "`${protocol}//${window.location.host}/ws`;" + `

        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            console.log('WebSocket连接已建立');
        };

        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            if (message.type === 'status_update') {
                this.updateUI(message.data);
            } else if (message.type === 'new_request') {
                this.addRequestLog(message.data);
            } else if (message.type === 'server_error') {
                this.handleServerError(message.error);
            }
        };

        this.ws.onclose = () => {
            console.log('WebSocket连接已关闭，尝试重连...');
            setTimeout(() => this.connectWebSocket(), 3000);
        };
    }

    bindEvents() {
        // 服务器控制
        document.getElementById('start-server').addEventListener('click', () => this.startServer());
        document.getElementById('stop-server').addEventListener('click', () => this.stopServer());

        // 发送请求
        document.getElementById('send-request').addEventListener('click', () => this.sendRequest());
        document.getElementById('format-json').addEventListener('click', () => this.formatJSON());

        // 日志管理
        document.getElementById('clear-logs').addEventListener('click', () => this.clearLogs());
        document.getElementById('refresh-logs').addEventListener('click', () => this.refreshLogs());

        // JSON编辑器
        document.getElementById('save-json').addEventListener('click', () => this.saveJSONFile());
        document.getElementById('close-json-editor').addEventListener('click', () => this.closeJSONEditor());
    }

    async loadStatus() {
        try {
            const response = await fetch('/api/status');
            const data = await response.json();
            this.updateUI(data);
        } catch (error) {
            console.error('加载状态失败:', error);
        }
    }

    async loadJSONFiles() {
        try {
            const response = await fetch('/api/files');
            const files = await response.json();
            this.jsonFiles = files || [];
            this.updateEndpointsUI();
        } catch (error) {
            console.error('加载JSON文件列表失败:', error);
            this.jsonFiles = [];
            this.updateEndpointsUI();
        }
    }

    updateUI(data) {
        // 更新服务器状态
        document.getElementById('server-ip').value = data.ip;
        document.getElementById('server-port').value = data.port;

        const statusElement = document.getElementById('server-status');
        const urlElement = document.getElementById('server-url');
        const startBtn = document.getElementById('start-server');
        const stopBtn = document.getElementById('stop-server');

        if (data.is_running) {
            statusElement.textContent = '运行中';
            statusElement.className = 'status-running';
            urlElement.textContent = ` + "`http://${data.ip}:${data.port}`;" + `
            startBtn.disabled = true;
            stopBtn.disabled = false;
        } else {
            statusElement.textContent = '已停止';
            statusElement.className = 'status-stopped';
            urlElement.textContent = '';
            startBtn.disabled = false;
            stopBtn.disabled = true;
        }

        // 更新接口配置
        this.endpoints = data.endpoints || [];
        this.updateEndpointsUI();
    }

    initTabs() {
        // 绑定TAB切换事件
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const targetTab = e.target.getAttribute('data-tab');
                this.switchTab(targetTab);
            });
        });
    }

    switchTab(tabName) {
        // 移除所有活动状态
        document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
        document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));

        // 激活目标TAB
        document.querySelector(` + "`[data-tab=\"${tabName}\"]`" + `).classList.add('active');
        document.getElementById(` + "`${tabName}-tab`" + `).classList.add('active');
    }

    updateEndpointsUI() {
        const container = document.getElementById('endpoints-config');
        container.innerHTML = '';

        this.endpoints.forEach((endpoint, index) => {
            const endpointDiv = document.createElement('div');
            endpointDiv.className = ` + "`endpoint-item ${endpoint.is_active ? 'active' : ''}`;" + `

            endpointDiv.innerHTML = ` + "`" + `
                <div class="endpoint-checkbox">
                    <input type="checkbox" id="endpoint-${index}" ${endpoint.is_active ? 'checked' : ''}
                           onchange="tool.toggleEndpoint(${index})">
                </div>
                <div class="endpoint-path">
                    <input type="text" value="${endpoint.path}"
                           onchange="tool.updateEndpointPath(${index}, this.value)"
                           placeholder="/api/path">
                </div>
                <div class="endpoint-file">
                    <select onchange="tool.updateEndpointFile(${index}, this.value)">
                        <option value="">响应文件</option>
                        ${this.jsonFiles.map(file =>
                            ` + "`<option value=\"${file}\" ${endpoint.response_file === file ? 'selected' : ''}>${file.substring(0, 12)}${file.length > 12 ? '...' : ''}</option>`" + `
                        ).join('')}
                    </select>
                    ${endpoint.response_file ? ` + "`<button class=\"btn-edit\" onclick=\"tool.editJSONFile('${endpoint.response_file}')\">编辑</button>`" + ` : ''}
                </div>
            ` + "`;" + `

            container.appendChild(endpointDiv);
        });
    }

    toggleEndpoint(index) {
        this.endpoints[index].is_active = !this.endpoints[index].is_active;
        this.updateEndpointsUI();
    }

    updateEndpointFile(index, file) {
        this.endpoints[index].response_file = file;
        this.updateEndpointsUI(); // 重新渲染UI以显示编辑按钮
    }

    updateEndpointPath(index, path) {
        if (!path.startsWith('/')) {
            path = '/' + path;
        }
        this.endpoints[index].path = path;
    }

    handleServerError(error) {
        const errorElement = document.getElementById('error-message');
        const startBtn = document.getElementById('start-server');
        const stopBtn = document.getElementById('stop-server');

        // 重置按钮状态
        startBtn.disabled = false;
        stopBtn.disabled = true;

        // 显示错误信息
        if (error.includes('bind') || error.includes('address already in use') ||
            error.includes('cannot assign requested address')) {
            errorElement.textContent = 'IP地址绑定失败，请检查IP地址是否正确或端口是否被占用';
        } else if (error.includes('permission denied')) {
            errorElement.textContent = '权限不足，请尝试使用其他端口或以管理员身份运行';
        } else {
            errorElement.textContent = '服务器启动失败: ' + error;
        }
        errorElement.style.display = 'inline';

        this.showMessage('服务器启动失败', 'error');
    }

    async startServer() {
        const errorElement = document.getElementById('error-message');
        errorElement.style.display = 'none';

        try {
            const response = await fetch('/api/start', {
                method: 'POST'
            });
            const result = await response.json();
            if (response.ok) {
                this.showMessage('服务器启动成功', 'success');
            } else {
                this.showMessage(result.error, 'error');
                if (result.error.includes('bind') || result.error.includes('address')) {
                    errorElement.textContent = 'IP绑定失败，请检查IP地址是否正确';
                    errorElement.style.display = 'inline';
                }
            }
        } catch (error) {
            this.showMessage('启动服务器失败: ' + error.message, 'error');
            errorElement.textContent = '网络连接错误';
            errorElement.style.display = 'inline';
        }
    }

    async stopServer() {
        try {
            const response = await fetch('/api/stop', {
                method: 'POST'
            });
            const result = await response.json();
            if (response.ok) {
                this.showMessage('服务器已停止', 'success');
            } else {
                this.showMessage(result.error, 'error');
            }
        } catch (error) {
            this.showMessage('停止服务器失败: ' + error.message, 'error');
        }
    }

    async saveConfig() {
        const config = {
            ip: document.getElementById('server-ip').value,
            port: document.getElementById('server-port').value,
            endpoints: this.endpoints
        };

        try {
            const response = await fetch('/api/config', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(config)
            });
            const result = await response.json();
            if (response.ok) {
                this.showMessage('配置保存成功', 'success');
            } else {
                this.showMessage(result.error, 'error');
            }
        } catch (error) {
            this.showMessage('保存配置失败: ' + error.message, 'error');
        }
    }

    async sendRequest() {
        const url = document.getElementById('request-url').value;
        const method = document.getElementById('request-method').value;
        const headersText = document.getElementById('request-headers').value;
        const data = document.getElementById('request-data').value;

        if (!url) {
            this.showMessage('请输入请求URL', 'error');
            return;
        }

        let headers = {};
        if (headersText) {
            try {
                headers = JSON.parse(headersText);
            } catch (error) {
                this.showMessage('请求头格式错误: ' + error.message, 'error');
                return;
            }
        }

        const request = {
            url: url,
            method: method,
            headers: headers,
            data: data
        };

        try {
            const response = await fetch('/api/send', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(request)
            });
            const result = await response.json();

            if (response.ok) {
                this.displayResponse(result);
            } else {
                this.showMessage('发送请求失败: ' + result.error, 'error');
            }
        } catch (error) {
            this.showMessage('发送请求失败: ' + error.message, 'error');
        }
    }

    displayResponse(response) {
        const resultElement = document.getElementById('response-result');

        let displayText = ` + "`状态码: ${response.status}\n\n`;" + `
        displayText += "响应头:\n";
        for (const [key, value] of Object.entries(response.headers)) {
            displayText += ` + "`${key}: ${value}\n`;" + `
        }
        displayText += "\n响应体:\n";

        // 尝试格式化JSON
        try {
            const jsonBody = JSON.parse(response.body);
            displayText += JSON.stringify(jsonBody, null, 2);
        } catch {
            displayText += response.body;
        }

        resultElement.textContent = displayText;
    }

    formatJSON() {
        const dataField = document.getElementById('request-data');

        // 格式化请求数据
        try {
            if (dataField.value.trim()) {
                const parsed = JSON.parse(dataField.value);
                dataField.value = JSON.stringify(parsed, null, 2);
            }
        } catch (error) {
            this.showMessage('请求数据JSON格式错误: ' + error.message, 'error');
        }
    }

    async refreshLogs() {
        try {
            const response = await fetch('/api/logs');
            const logs = await response.json();
            this.displayLogs(logs);
        } catch (error) {
            this.showMessage('刷新日志失败: ' + error.message, 'error');
        }
    }

    clearLogs() {
        document.getElementById('request-logs').innerHTML = '<p>日志已清空</p>';
    }

    displayLogs(logs) {
        const container = document.getElementById('request-logs');
        container.innerHTML = '';

        if (logs.length === 0) {
            container.innerHTML = '<p>暂无请求日志</p>';
            return;
        }

        logs.reverse().forEach((log, index) => {
            this.addRequestLog(log, index === 0);
        });
    }

    addRequestLog(log, prepend = true) {
        const container = document.getElementById('request-logs');
        const logDiv = document.createElement('div');
        logDiv.className = 'log-item';

        const timestamp = new Date(log.timestamp).toLocaleString('zh-CN', {
            month: '2-digit', day: '2-digit',
            hour: '2-digit', minute: '2-digit', second: '2-digit'
        });

        // 提取JSON内容用于显示
        let bodyPreview = log.body || '';
        let fullBody = log.body || '';
        if (bodyPreview.length > 100) {
            bodyPreview = bodyPreview.substring(0, 100) + '...';
        }

        // 尝试格式化JSON用于完整显示
        let formattedFullBody = fullBody;
        try {
            if (fullBody.trim()) {
                const parsed = JSON.parse(fullBody);
                formattedFullBody = JSON.stringify(parsed, null, 2);
            }
        } catch (e) {
            // 如果不是有效JSON，保持原样
        }

        logDiv.innerHTML = ` + "`" + `
            <div class="log-header">
                <span class="log-method ${log.method}">${log.method}</span>
                <span>${log.path}</span>
                <span style="font-size: 11px; color: #666;">${timestamp}</span>
                <span style="font-size: 11px; color: #666;">${Object.keys(log.headers).length}个头</span>
                <button class="toggle-btn" onclick="this.parentElement.nextElementSibling.style.display = this.parentElement.nextElementSibling.style.display === 'none' ? 'block' : 'none'">展开</button>
            </div>
            <div class="log-content" style="display: none;">
                <div><strong>请求体:</strong></div>
                <pre style="background: #f5f5f5; padding: 8px; border-radius: 3px; margin: 5px 0; white-space: pre-wrap; font-family: monospace; font-size: 12px; max-height: 300px; overflow-y: auto;">${formattedFullBody || '(空)'}</pre>
                ${Object.keys(log.headers).length > 0 ? ` + "`<div style=\"margin-top: 5px;\"><strong>主要请求头:</strong></div><div>${Object.entries(log.headers).slice(0, 3).map(([k,v]) => `${k}: ${Array.isArray(v) ? v[0] : v}`).join('<br>')}</div>`" + ` : ''}
            </div>
        ` + "`;" + `

        if (prepend && container.firstChild) {
            container.insertBefore(logDiv, container.firstChild);
        } else {
            container.appendChild(logDiv);
        }

        // 限制显示的日志数量
        const logItems = container.querySelectorAll('.log-item');
        if (logItems.length > 50) {
            container.removeChild(container.lastChild);
        }
    }

    async editJSONFile(filename) {
        try {
            const response = await fetch(` + "`/json_files/${filename}`" + `);
            if (response.ok) {
                const content = await response.text();
                document.getElementById('json-file-name').textContent = ` + "`编辑: ${filename}`;" + `
                document.getElementById('json-content').value = content;
                document.getElementById('json-editor').style.display = 'block';
                this.currentEditingFile = filename;

                // 格式化JSON显示
                try {
                    const parsed = JSON.parse(content);
                    document.getElementById('json-content').value = JSON.stringify(parsed, null, 2);
                } catch (e) {
                    // 如果不是有效JSON，保持原样
                }
            } else {
                this.showMessage('无法读取文件', 'error');
            }
        } catch (error) {
            this.showMessage('读取文件失败: ' + error.message, 'error');
        }
    }

    async saveJSONFile() {
        if (!this.currentEditingFile) {
            this.showMessage('没有正在编辑的文件', 'error');
            return;
        }

        const content = document.getElementById('json-content').value;

        // 验证JSON格式
        try {
            JSON.parse(content);
        } catch (error) {
            this.showMessage('JSON格式错误: ' + error.message, 'error');
            return;
        }

        try {
            const response = await fetch('/api/save-json', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    filename: this.currentEditingFile,
                    content: content
                })
            });

            if (response.ok) {
                this.showMessage('文件保存成功', 'success');
            } else {
                const result = await response.json();
                this.showMessage('保存失败: ' + result.error, 'error');
            }
        } catch (error) {
            this.showMessage('保存失败: ' + error.message, 'error');
        }
    }

    closeJSONEditor() {
        document.getElementById('json-editor').style.display = 'none';
        this.currentEditingFile = null;
    }

    showMessage(message, type = 'info') {
        // 创建消息提示
        const messageDiv = document.createElement('div');
        messageDiv.style.cssText = ` + "`" + `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 15px 20px;
            border-radius: 5px;
            color: white;
            font-weight: bold;
            z-index: 1000;
            max-width: 300px;
            word-wrap: break-word;
        ` + "`;" + `

        if (type === 'success') {
            messageDiv.style.backgroundColor = '#27ae60';
        } else if (type === 'error') {
            messageDiv.style.backgroundColor = '#e74c3c';
        } else {
            messageDiv.style.backgroundColor = '#3498db';
        }

        messageDiv.textContent = message;
        document.body.appendChild(messageDiv);

        setTimeout(() => {
            document.body.removeChild(messageDiv);
        }, 3000);
    }

    // 加载项目列表
    async loadProjects() {
        try {
            const response = await fetch('/api/projects');
            const projects = await response.json();

            // 获取当前项目
            const statusResponse = await fetch('/api/status');
            const status = await statusResponse.json();
            const currentProject = status.current_project || 'default';

            const select = document.getElementById('project-select');
            select.innerHTML = '';

            projects.forEach(proj => {
                const option = document.createElement('option');
                option.value = proj.name;
                option.textContent = proj.name;
                // 设置当前项目为选中状态
                if (proj.name === currentProject) {
                    option.selected = true;
                }
                select.appendChild(option);
            });
        } catch (error) {
            console.error('加载项目列表失败:', error);
        }
    }
}

// 全局函数：切换项目
async function switchProject() {
    const select = document.getElementById('project-select');
    const projectName = select.value;

    try {
        const response = await fetch('/api/switch-project', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ project: projectName })
        });

        if (response.ok) {
            // 重新加载状态和文件列表
            window.location.reload();
        } else {
            const error = await response.json();
            alert('切换项目失败: ' + error.error);
        }
    } catch (error) {
        alert('切换项目失败: ' + error.message);
    }
}

// 全局函数：创建新项目
async function createNewProject() {
    const projectName = prompt('请输入项目名称:');
    if (!projectName) return;

    // 验证项目名
    if (projectName.includes('/') || projectName.includes('\\') || projectName.includes('..')) {
        alert('项目名称包含非法字符！');
        return;
    }

    try {
        const response = await fetch('/api/projects', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name: projectName })
        });

        if (response.ok) {
            alert('项目创建成功！');
            // 重新加载项目列表
            await tool.loadProjects();
            // 切换到新项目
            document.getElementById('project-select').value = projectName;
            await switchProject();
        } else {
            const error = await response.json();
            alert('项目创建失败: ' + error.error);
        }
    } catch (error) {
        alert('项目创建失败: ' + error.message);
    }
}

// 全局函数：保存项目配置
async function saveProjectConfig() {
    try {
        // 调用工具类的保存配置方法
        await tool.saveConfig();
        alert('项目配置保存成功！');
    } catch (error) {
        alert('保存项目配置失败: ' + error.message);
    }
}

// 初始化应用
const tool = new HTTPJSONTool();`

	os.WriteFile("static/app.js", []byte(js), 0644)
}

func createSampleJSONFiles() {
	// 创建示例JSON文件
	sampleFiles := map[string]string{
		"success_response.json": `{
    "code": 200,
    "message": "操作成功",
    "data": {
        "id": 12345,
        "name": "示例数据",
        "timestamp": "2024-01-01T12:00:00Z"
    }
}`,
		"error_response.json": `{
    "code": 500,
    "message": "内部服务器错误",
    "error": "处理请求时发生错误"
}`,
		"audit_task_result.json": `{
    "taskId": "audio_task_001",
    "status": "completed",
    "result": {
        "duration": 120.5,
        "quality": "excellent",
        "transcription": "这是一个音频任务的审核结果示例",
        "confidence": 0.95
    },
    "timestamp": "2024-01-01T14:30:00Z"
}`,
		"user_list.json": `{
    "users": [
        {
            "id": 1,
            "username": "admin",
            "email": "admin@example.com",
            "role": "administrator"
        },
        {
            "id": 2,
            "username": "user1",
            "email": "user1@example.com",
            "role": "user"
        }
    ],
    "total": 2,
    "page": 1,
    "pageSize": 10
}`,
		"config_data.json": `{
    "server": {
        "host": "localhost",
        "port": 8080,
        "ssl": false
    },
    "database": {
        "driver": "mysql",
        "host": "localhost",
        "port": 3306,
        "name": "test_db"
    },
    "features": {
        "logging": true,
        "cache": true,
        "debug": false
    }
}`,
	}

	// 在默认项目目录下创建示例文件
	jsonFilesPath := filepath.Join("projects", "default", "json_files")
	for filename, content := range sampleFiles {
		filePath := filepath.Join(jsonFilesPath, filename)
		// 如果文件不存在才创建
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			os.WriteFile(filePath, []byte(content), 0644)
		}
	}
}