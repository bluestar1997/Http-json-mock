class HTTPJSONTool {
    constructor() {
        this.ws = null;
        this.init();
        this.connectWebSocket();
    }

    async init() {
        this.sendBlocks = [];
        this.bindEvents();
        this.loadProjects();
        await this.loadJSONFiles();
        this.loadStatus();
        this.initTabs();
        this.loadSendBlocks();
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

        // 日志管理
        document.getElementById('clear-logs').addEventListener('click', () => this.clearLogs());
        document.getElementById('refresh-logs').addEventListener('click', () => this.refreshLogs());

        // JSON编辑器
        document.getElementById('edit-json').addEventListener('click', () => this.enableJSONEdit());
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
        } catch (error) {
            console.error('加载JSON文件列表失败:', error);
            this.jsonFiles = [];
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
            endpointDiv.className = 'endpoint-item';
            endpointDiv.style.marginBottom = '5px';

            endpointDiv.innerHTML = `
                <div style="margin-bottom: 3px;">
                    <input type="text" value="${endpoint.name || ''}"
                           onchange="tool.updateEndpointName(${index}, this.value)"
                           placeholder="接口名称"
                           style="width: 100%; padding: 4px; font-size: 13px;">
                </div>
                <div style="display: flex; gap: 8px;">
                    <div style="flex: 0 0 60%;">
                        <input type="text" value="${endpoint.path}"
                               onchange="tool.updateEndpointPath(${index}, this.value)"
                               placeholder="/api/path"
                               style="width: 100%; padding: 4px; font-size: 13px;">
                    </div>
                    <div style="flex: 0 0 40%;">
                        <select onchange="tool.selectEndpointFile(${index}, this.value)"
                                style="width: 100%; padding: 4px; font-size: 13px;">
                            <option value="">响应文件</option>
                            ${this.jsonFiles.map(file =>
                                `<option value="${file}" ${endpoint.response_file === file ? 'selected' : ''}>${file}</option>`
                            ).join('')}
                        </select>
                    </div>
                </div>
            `;

            container.appendChild(endpointDiv);
        });
    }

    updateEndpointName(index, name) {
        this.endpoints[index].name = name;
    }

    updateEndpointFile(index, file) {
        this.endpoints[index].response_file = file;
    }

    // 选择接口文件并自动加载到编辑区
    async selectEndpointFile(index, file) {
        this.endpoints[index].response_file = file;

        if (file) {
            // 自动打开编辑器并加载文件
            await this.editJSONFile(file);
        }
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
        // 先更新发送块配置
        await this.saveSendBlocksConfig();

        const config = {
            ip: document.getElementById('server-ip').value,
            port: document.getElementById('server-port').value,
            endpoints: this.endpoints,
            send_blocks: this.sendBlocks
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

    // 加载发送块配置
    async loadSendBlocks() {
        try {
            const response = await fetch('/api/status');
            const data = await response.json();

            // 如果配置中有send_blocks，使用配置的；否则使用默认的2个块
            if (data.send_blocks && data.send_blocks.length > 0) {
                this.sendBlocks = data.send_blocks;
            } else {
                this.sendBlocks = [
                    { name: '', url: '', send_file: '', method: 'POST', headers: '{"content-type":"application/json"}' },
                    { name: '', url: '', send_file: '', method: 'POST', headers: '{"content-type":"application/json"}' }
                ];
            }
            this.renderSendBlocks();
        } catch (error) {
            console.error('加载发送块配置失败:', error);
            // 使用默认配置
            this.sendBlocks = [
                { name: '', url: '', send_file: '', method: 'POST', headers: '{"content-type":"application/json"}' },
                { name: '', url: '', send_file: '', method: 'POST', headers: '{"content-type":"application/json"}' }
            ];
            this.renderSendBlocks();
        }
    }

    // 渲染所有发送块
    renderSendBlocks() {
        const container = document.getElementById('send-blocks-container');
        container.innerHTML = '';

        this.sendBlocks.forEach((block, index) => {
            container.appendChild(this.createSendBlockElement(block, index));
        });
    }

    // 创建单个发送块元素
    createSendBlockElement(block, index) {
        const section = document.createElement('section');
        section.className = 'section compact send-block';
        section.style.marginBottom = '20px';
        section.style.border = '1px solid #ddd';
        section.style.borderRadius = '4px';
        section.style.padding = '15px';

        const html = `
            <div style="margin-bottom: 10px;">
                <input type="text" id="send-name-${index}" value="${block.name || ''}" placeholder="请输入发送功能描述" style="width: 100%; padding: 8px; font-size: 14px; font-weight: bold; border: 1px solid #ddd; border-radius: 4px;">
            </div>
            <div class="send-block-form">
                <div class="form-row" style="display: flex; gap: 10px; margin-bottom: 10px; align-items: flex-end;">
                    <div class="form-group" style="flex: 0 0 100px;">
                        <select id="send-method-${index}" style="width: 100%; padding: 8px;">
                            <option value="POST" ${block.method === 'POST' ? 'selected' : ''}>POST</option>
                            <option value="GET" ${block.method === 'GET' ? 'selected' : ''}>GET</option>
                        </select>
                    </div>
                    <div class="form-group" style="flex: 1;">
                        <input type="text" id="send-url-${index}" value="${block.url || ''}" placeholder="http://..." style="width: 100%; padding: 8px;">
                    </div>
                    <div class="form-group" style="flex: 0 0 180px;">
                        <select id="send-file-${index}" onchange="tool.loadSendFile(${index})" style="width: 100%; padding: 8px;">
                            <option value="">选择JSON文件</option>
                        </select>
                    </div>
                    <div class="form-group" style="flex: 0 0 260px;">
                        <input type="text" id="send-headers-${index}" value='${block.headers || '{"content-type":"application/json"}'}' placeholder='{"content-type":"application/json"}' style="width: 100%; padding: 8px;">
                    </div>
                    <div class="form-actions" style="display: flex; gap: 5px;">
                        <button onclick="tool.sendBlockRequest(${index})" class="btn btn-primary" style="padding: 8px 20px;">发送</button>
                    </div>
                </div>
                <div class="form-row" style="display: flex; gap: 10px; margin-bottom: 10px;">
                    <div class="form-group" style="flex: 1;">
                        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 5px;">
                            <label style="margin: 0;">请求数据 (JSON):</label>
                            <div style="display: flex; gap: 5px;">
                                <button id="edit-btn-${index}" onclick="tool.enableEditMode(${index})" class="btn btn-info" style="padding: 3px 10px;">编辑</button>
                                <button id="format-btn-${index}" onclick="tool.formatSendBlockJSON(${index})" class="btn btn-info" style="padding: 3px 10px; display: none;">格式化</button>
                                <button id="save-btn-${index}" onclick="tool.saveSendBlockData(${index})" class="btn btn-success" style="padding: 3px 10px; display: none;">保存</button>
                            </div>
                        </div>
                        <textarea id="send-data-${index}" rows="6" placeholder='{"key": "value"}' style="width: 100%; background-color: #f5f5f5;" readonly>${block.data || ''}</textarea>
                    </div>
                </div>
                <div class="response-section">
                    <h4>响应结果:</h4>
                    <pre id="send-response-${index}" style="background: #f5f5f5; padding: 10px; border-radius: 4px; max-height: 300px; overflow: auto;"></pre>
                </div>
            </div>
        `;

        section.innerHTML = html;

        // 填充文件列表
        setTimeout(() => {
            const fileSelect = document.getElementById(`send-file-${index}`);
            if (fileSelect && this.jsonFiles) {
                this.jsonFiles.forEach(file => {
                    const option = document.createElement('option');
                    option.value = file;
                    option.textContent = file;
                    if (file === block.send_file) {
                        option.selected = true;
                    }
                    fileSelect.appendChild(option);
                });
            }
        }, 100);

        return section;
    }

    // 添加新的发送块
    addSendBlock() {
        this.sendBlocks.push({
            name: '',
            url: '',
            send_file: '',
            method: 'POST',
            headers: '{"content-type":"application/json"}'
        });
        this.renderSendBlocks();
        this.saveSendBlocksConfig();
    }

    // 发送块的请求
    async sendBlockRequest(index) {
        const url = document.getElementById(`send-url-${index}`).value;
        const method = document.getElementById(`send-method-${index}`).value;
        const headersText = document.getElementById(`send-headers-${index}`).value;
        const data = document.getElementById(`send-data-${index}`).value;

        if (!url) {
            alert('请输入请求URL');
            return;
        }

        let headers = {};
        if (headersText) {
            try {
                headers = JSON.parse(headersText);
            } catch (error) {
                alert('请求头格式错误: ' + error.message);
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
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(request)
            });
            const result = await response.json();

            const responseElement = document.getElementById(`send-response-${index}`);
            if (response.ok) {
                let displayText = `状态码: ${result.status}

`;
                displayText += "响应头:\n";
                for (const [key, value] of Object.entries(result.headers)) {
                    displayText += `${key}: ${value}
`;
                }
                displayText += "\n响应体:\n";
                try {
                    const jsonBody = JSON.parse(result.body);
                    displayText += JSON.stringify(jsonBody, null, 2);
                } catch {
                    displayText += result.body;
                }
                responseElement.textContent = displayText;
            } else {
                responseElement.textContent = '发送请求失败: ' + result.error;
            }
        } catch (error) {
            document.getElementById(`send-response-${index}`).textContent = '发送请求失败: ' + error.message;
        }

        // 保存当前配置
        this.updateSendBlockConfig(index);
    }

    // 启用编辑模式
    async enableEditMode(index) {
        const dataField = document.getElementById(`send-data-${index}`);
        const sendFileSelect = document.getElementById(`send-file-${index}`);
        const editBtn = document.getElementById(`edit-btn-${index}`);
        const formatBtn = document.getElementById(`format-btn-${index}`);
        const saveBtn = document.getElementById(`save-btn-${index}`);

        // 如果请求数据为空且选择了发送文件，则自动加载文件内容
        if (!dataField.value.trim() && sendFileSelect.value) {
            try {
                const response = await fetch(`/api/read-json?file=${encodeURIComponent(sendFileSelect.value)}`);
                if (response.ok) {
                    const result = await response.json();
                    dataField.value = JSON.stringify(result, null, 2);
                }
            } catch (error) {
                console.error('加载文件失败:', error);
            }
        }

        // 移除只读属性
        dataField.removeAttribute('readonly');
        dataField.style.backgroundColor = '#fff';

        // 切换按钮显示
        editBtn.style.display = 'none';
        formatBtn.style.display = 'inline-block';
        saveBtn.style.display = 'inline-block';
    }

    // 禁用编辑模式
    disableEditMode(index) {
        const dataField = document.getElementById(`send-data-${index}`);
        const editBtn = document.getElementById(`edit-btn-${index}`);
        const formatBtn = document.getElementById(`format-btn-${index}`);
        const saveBtn = document.getElementById(`save-btn-${index}`);

        // 添加只读属性
        dataField.setAttribute('readonly', 'readonly');
        dataField.style.backgroundColor = '#f5f5f5';

        // 切换按钮显示
        editBtn.style.display = 'inline-block';
        formatBtn.style.display = 'none';
        saveBtn.style.display = 'none';
    }

    // 格式化发送块的JSON
    formatSendBlockJSON(index) {
        const dataField = document.getElementById(`send-data-${index}`);
        try {
            if (dataField.value.trim()) {
                const parsed = JSON.parse(dataField.value);
                dataField.value = JSON.stringify(parsed, null, 2);
            }
        } catch (error) {
            alert('JSON格式错误: ' + error.message);
        }
    }

    // 保存发送块的数据到文件
    async saveSendBlockData(index) {
        const sendFile = document.getElementById(`send-file-${index}`).value;
        const data = document.getElementById(`send-data-${index}`).value;

        if (!sendFile) {
            alert('请先选择发送文件');
            return;
        }

        if (!data.trim()) {
            alert('请求数据为空');
            return;
        }

        // 验证JSON格式
        try {
            JSON.parse(data);
        } catch (error) {
            alert('JSON格式错误，无法保存: ' + error.message);
            return;
        }

        try {
            const response = await fetch('/api/save-json', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    filename: sendFile,
                    content: data
                })
            });

            if (response.ok) {
                alert('保存成功');
                // 保存成功后退出编辑模式
                this.disableEditMode(index);
            } else {
                const result = await response.json();
                alert('保存失败: ' + result.error);
            }
        } catch (error) {
            alert('保存失败: ' + error.message);
        }
    }

    // 加载发送文件内容
    async loadSendFile(index) {
        const sendFile = document.getElementById(`send-file-${index}`).value;

        if (!sendFile) {
            document.getElementById(`send-data-${index}`).value = '';
            return;
        }

        try {
            const response = await fetch(`/api/read-json?file=${encodeURIComponent(sendFile)}`);
            if (response.ok) {
                const result = await response.json();
                document.getElementById(`send-data-${index}`).value = JSON.stringify(result, null, 2);
                // 加载文件后自动进入编辑模式
                this.enableEditMode(index);
            } else {
                alert('读取文件失败');
            }
        } catch (error) {
            console.error('读取文件失败:', error);
        }

        // 更新配置
        this.updateSendBlockConfig(index);
    }

    // 更新单个发送块配置
    updateSendBlockConfig(index) {
        if (index >= 0 && index < this.sendBlocks.length) {
            this.sendBlocks[index] = {
                name: document.getElementById(`send-name-${index}`).value,
                url: document.getElementById(`send-url-${index}`).value,
                send_file: document.getElementById(`send-file-${index}`).value,
                method: document.getElementById(`send-method-${index}`).value,
                headers: document.getElementById(`send-headers-${index}`).value
            };
            this.saveSendBlocksConfig();
        }
    }

    // 保存发送块配置到服务器
    async saveSendBlocksConfig() {
        // 收集所有发送块的当前配置
        const blocks = [];
        for (let i = 0; i < this.sendBlocks.length; i++) {
            const nameElem = document.getElementById(`send-name-${i}`);
            const urlElem = document.getElementById(`send-url-${i}`);
            const fileElem = document.getElementById(`send-file-${i}`);
            const methodElem = document.getElementById(`send-method-${i}`);
            const headersElem = document.getElementById(`send-headers-${i}`);

            if (nameElem && urlElem && fileElem && methodElem && headersElem) {
                blocks.push({
                    name: nameElem.value,
                    url: urlElem.value,
                    send_file: fileElem.value,
                    method: methodElem.value,
                    headers: headersElem.value
                });
            }
        }

        this.sendBlocks = blocks;

        // 这个配置会在saveProjectConfig时一起保存
        console.log('发送块配置已更新', this.sendBlocks);
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
            const response = await fetch(`/api/read-json?file=${encodeURIComponent(filename)}`);
            if (response.ok) {
                const jsonData = await response.json();
                const textarea = document.getElementById('json-content');
                const editBtn = document.getElementById('edit-json');
                const saveBtn = document.getElementById('save-json');

                document.getElementById('json-file-name').textContent = `编辑: ${filename}`;
                // 格式化JSON显示
                textarea.value = JSON.stringify(jsonData, null, 2);
                document.getElementById('json-editor').style.display = 'block';
                this.currentEditingFile = filename;

                // 重置为只读模式
                textarea.setAttribute('readonly', 'readonly');
                textarea.style.backgroundColor = '#f5f5f5';
                editBtn.style.display = 'inline-block';
                saveBtn.style.display = 'none';
            } else {
                const error = await response.json();
                this.showMessage('无法读取文件: ' + (error.error || '未知错误'), 'error');
            }
        } catch (error) {
            this.showMessage('读取文件失败: ' + error.message, 'error');
        }
    }

    // 启用JSON编辑模式
    enableJSONEdit() {
        const textarea = document.getElementById('json-content');
        const editBtn = document.getElementById('edit-json');
        const saveBtn = document.getElementById('save-json');

        // 移除只读属性
        textarea.removeAttribute('readonly');
        textarea.style.backgroundColor = '#fff';

        // 切换按钮显示
        editBtn.style.display = 'none';
        saveBtn.style.display = 'inline-block';
    }

    // 禁用JSON编辑模式
    disableJSONEdit() {
        const textarea = document.getElementById('json-content');
        const editBtn = document.getElementById('edit-json');
        const saveBtn = document.getElementById('save-json');

        // 添加只读属性
        textarea.setAttribute('readonly', 'readonly');
        textarea.style.backgroundColor = '#f5f5f5';

        // 切换按钮显示
        editBtn.style.display = 'inline-block';
        saveBtn.style.display = 'none';
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
                // 保存成功后退出编辑模式
                this.disableJSONEdit();
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

// 全局函数：添加发送块
function addSendBlock() {
    tool.addSendBlock();
}

// 初始化应用
const tool = new HTTPJSONTool();