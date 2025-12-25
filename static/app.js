class HTTPJSONTool {
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
        const wsUrl = `${protocol}//${window.location.host}/ws`;

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
            urlElement.textContent = `http://${data.ip}:${data.port}`;
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
        document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');
        document.getElementById(`${tabName}-tab`).classList.add('active');
    }

    updateEndpointsUI() {
        const container = document.getElementById('endpoints-config');
        container.innerHTML = '';

        this.endpoints.forEach((endpoint, index) => {
            const endpointDiv = document.createElement('div');
            endpointDiv.className = `endpoint-item ${endpoint.is_active ? 'active' : ''}`;

            endpointDiv.innerHTML = `
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
                            `<option value="${file}" ${endpoint.response_file === file ? 'selected' : ''}>${file.substring(0, 12)}${file.length > 12 ? '...' : ''}</option>`
                        ).join('')}
                    </select>
                    ${endpoint.response_file ? `<button class="btn-edit" onclick="tool.editJSONFile('${endpoint.response_file}')">编辑</button>` : ''}
                </div>
            `;

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

        let displayText = `状态码: ${response.status}

`;
        displayText += "响应头:\n";
        for (const [key, value] of Object.entries(response.headers)) {
            displayText += `${key}: ${value}
`;
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

        logDiv.innerHTML = `
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
                ${Object.keys(log.headers).length > 0 ? `<div style="margin-top: 5px;"><strong>主要请求头:</strong></div><div>${Object.entries(log.headers).slice(0, 3).map(([k,v]) => `${k}: ${Array.isArray(v) ? v[0] : v}`).join('<br>')}</div>` : ''}
            </div>
        `;

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
            const response = await fetch(`/json_files/${filename}`);
            if (response.ok) {
                const content = await response.text();
                document.getElementById('json-file-name').textContent = `编辑: ${filename}`;
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
        messageDiv.style.cssText = `
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
        `;

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
const tool = new HTTPJSONTool();