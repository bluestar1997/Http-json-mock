package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Server struct {
	IP                string                 `json:"ip"`
	Port              string                 `json:"port"`
	IsRunning         bool                   `json:"is_running"`
	Endpoints         []EndpointConfig       `json:"endpoints"`
	SendBlocks        []SendBlock            `json:"send_blocks"`
	RequestLogs       []RequestLog           `json:"request_logs"`
	mu                sync.RWMutex
	engine            *gin.Engine
	upgrader          websocket.Upgrader
	clients           map[*websocket.Conn]bool
	clientsMu         sync.RWMutex
}

type EndpointConfig struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	ResponseFile string `json:"response_file"`
}

type RequestLog struct {
	ID        int                    `json:"id"`
	Path      string                 `json:"path"`
	Method    string                 `json:"method"`
	Headers   map[string]interface{} `json:"headers"`
	Body      string                 `json:"body"`
	Timestamp time.Time              `json:"timestamp"`
}

type SendRequest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Data    string            `json:"data"`
}

type Config struct {
	IP             string           `json:"ip"`
	Port           string           `json:"port"`
	CurrentProject string           `json:"current_project"`
	Endpoints      []EndpointConfig `json:"endpoints"`
	SendBlocks     []SendBlock      `json:"send_blocks"`
}

type SendBlock struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	SendFile string `json:"send_file"`
	Method   string `json:"method"`
	Headers  string `json:"headers"`
}

type ProjectInfo struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	CreatedAt string `json:"created_at"`
}

var server *Server
var currentProject = "default"

func init() {
	server = &Server{
		IP:        "0.0.0.0",
		Port:      "29800",
		IsRunning: false,
		Endpoints: []EndpointConfig{
			{Name: "音频任务审计结果", Path: "/api/audioTask/getAuditTaskResult", ResponseFile: ""},
			{Name: "测试接口1", Path: "/api/test1", ResponseFile: ""},
			{Name: "测试接口2", Path: "/api/test2", ResponseFile: ""},
			{Name: "测试接口3", Path: "/api/test3", ResponseFile: ""},
		},
		SendBlocks:  []SendBlock{},
		RequestLogs: []RequestLog{},
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients: make(map[*websocket.Conn]bool),
	}
}

func main() {
	// 初始化项目结构
	if err := initializeProjects(); err != nil {
		log.Printf("初始化项目结构失败: %v", err)
	}

	// 创建必要的目录和文件
	os.MkdirAll("templates", 0755)
	os.MkdirAll("static", 0755)

	// 加载全局配置文件
	if err := loadGlobalConfig(); err != nil {
		log.Printf("加载全局配置文件失败: %v", err)
	}

	// 加载当前项目配置
	if err := loadConfig(); err != nil {
		log.Printf("加载项目配置文件失败: %v", err)
	}

	createHTMLTemplate()
	createCSSFile()
	createJSFile()
	createSampleJSONFiles()

	r := gin.Default()

	// 静态文件服务
	r.Static("/static", "./static")
	r.Static("/json_files", getJSONFilesPath(currentProject))
	r.LoadHTMLGlob("templates/*")

	// 主页
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// WebSocket连接
	r.GET("/ws", handleWebSocket)

	// API路由
	api := r.Group("/api")
	{
		api.GET("/status", getStatus)
		api.POST("/start", startServer)
		api.POST("/stop", stopServer)
		api.POST("/config", updateConfig)
		api.GET("/logs", getLogs)
		api.POST("/send", sendRequest)
		api.GET("/files", listJSONFiles)
		api.POST("/save-json", saveJSONFile)
		api.GET("/read-json", readJSONFile)
		api.GET("/projects", listProjects)
		api.POST("/projects", createProject)
		api.POST("/switch-project", switchProject)
	}

	log.Println("HTTP+JSON工具启动在 http://localhost:8080")
	r.Run(":8080")
}

func handleWebSocket(c *gin.Context) {
	conn, err := server.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket升级失败:", err)
		return
	}
	defer conn.Close()

	server.clientsMu.Lock()
	server.clients[conn] = true
	server.clientsMu.Unlock()

	defer func() {
		server.clientsMu.Lock()
		delete(server.clients, conn)
		server.clientsMu.Unlock()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

var broadcastMu sync.Mutex

func broadcastToClients(data interface{}) {
	broadcastMu.Lock()
	defer broadcastMu.Unlock()

	server.clientsMu.RLock()
	clients := make([]*websocket.Conn, 0, len(server.clients))
	for client := range server.clients {
		clients = append(clients, client)
	}
	server.clientsMu.RUnlock()

	message, _ := json.Marshal(data)
	var toRemove []*websocket.Conn

	for _, client := range clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			client.Close()
			toRemove = append(toRemove, client)
		}
	}

	// 移除断开的连接
	if len(toRemove) > 0 {
		server.clientsMu.Lock()
		for _, client := range toRemove {
			delete(server.clients, client)
		}
		server.clientsMu.Unlock()
	}
}

func getStatus(c *gin.Context) {
	server.mu.RLock()
	defer server.mu.RUnlock()

	response := map[string]interface{}{
		"ip":              server.IP,
		"port":            server.Port,
		"is_running":      server.IsRunning,
		"endpoints":       server.Endpoints,
		"send_blocks":     server.SendBlocks,
		"request_logs":    server.RequestLogs,
		"current_project": currentProject,
	}

	c.JSON(http.StatusOK, response)
}

func startServer(c *gin.Context) {
	server.mu.Lock()
	defer server.mu.Unlock()

	if server.IsRunning {
		c.JSON(http.StatusBadRequest, gin.H{"error": "服务器已在运行"})
		return
	}

	// 先设置为运行状态，防止重复启动
	server.IsRunning = true

	server.engine = gin.New()
	server.engine.Use(gin.Logger(), gin.Recovery())

	// 设置动态路由处理
	for _, endpoint := range server.Endpoints {
		path := endpoint.Path
		responseFile := endpoint.ResponseFile

		server.engine.Any(path, func(c *gin.Context) {
			handleDynamicEndpoint(c, path, responseFile)
		})
	}

	// 在单独的goroutine中启动服务器
	go func() {
		addr := fmt.Sprintf("%s:%s", server.IP, server.Port)
		log.Printf("HTTP服务器启动在 http://%s", addr)
		err := server.engine.Run(addr)

		if err != nil {
			log.Printf("服务器启动失败: %v", err)
			server.mu.Lock()
			server.IsRunning = false
			server.mu.Unlock()

			// 通知前端启动失败
			broadcastToClients(map[string]interface{}{
				"type": "server_error",
				"error": err.Error(),
			})
			broadcastToClients(map[string]interface{}{"type": "status_update", "data": server})
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "服务器启动中..."})
	broadcastToClients(map[string]interface{}{"type": "status_update", "data": server})
}

func stopServer(c *gin.Context) {
	server.mu.Lock()
	defer server.mu.Unlock()

	if !server.IsRunning {
		c.JSON(http.StatusBadRequest, gin.H{"error": "服务器未运行"})
		return
	}

	server.IsRunning = false
	server.engine = nil

	c.JSON(http.StatusOK, gin.H{"message": "服务器已停止"})
	broadcastToClients(map[string]interface{}{"type": "status_update", "data": server})
}

func updateConfig(c *gin.Context) {
	var config struct {
		IP         string           `json:"ip"`
		Port       string           `json:"port"`
		Endpoints  []EndpointConfig `json:"endpoints"`
		SendBlocks []SendBlock      `json:"send_blocks"`
	}

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	server.mu.Lock()
	server.IP = config.IP
	server.Port = config.Port
	server.Endpoints = config.Endpoints
	server.SendBlocks = config.SendBlocks
	server.mu.Unlock()

	// 保存配置到文件
	if err := saveConfig(); err != nil {
		log.Printf("保存配置文件失败: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置更新成功"})
	broadcastToClients(map[string]interface{}{"type": "status_update", "data": server})
}

func handleDynamicEndpoint(c *gin.Context, path, responseFile string) {
	// 记录请求
	headers := make(map[string]interface{})
	for k, v := range c.Request.Header {
		headers[k] = v
	}

	body, _ := io.ReadAll(c.Request.Body)

	requestLog := RequestLog{
		ID:        len(server.RequestLogs) + 1,
		Path:      path,
		Method:    c.Request.Method,
		Headers:   headers,
		Body:      string(body),
		Timestamp: time.Now(),
	}

	server.mu.Lock()
	server.RequestLogs = append(server.RequestLogs, requestLog)
	// 只保留最新的100条记录
	if len(server.RequestLogs) > 100 {
		server.RequestLogs = server.RequestLogs[1:]
	}
	server.mu.Unlock()

	// 广播新的请求日志
	broadcastToClients(map[string]interface{}{"type": "new_request", "data": requestLog})

	// 返回响应数据
	if responseFile != "" {
		data, err := os.ReadFile(filepath.Join(getJSONFilesPath(currentProject), responseFile))
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"message": "默认响应", "timestamp": time.Now()})
		} else {
			var jsonData interface{}
			if json.Unmarshal(data, &jsonData) == nil {
				c.JSON(http.StatusOK, jsonData)
			} else {
				c.String(http.StatusOK, string(data))
			}
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "默认响应", "timestamp": time.Now()})
	}
}

func getLogs(c *gin.Context) {
	server.mu.RLock()
	defer server.mu.RUnlock()

	c.JSON(http.StatusOK, server.RequestLogs)
}

func sendRequest(c *gin.Context) {
	var req SendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建HTTP请求
	var reqBody io.Reader
	if req.Data != "" {
		reqBody = strings.NewReader(req.Data)
	}

	httpReq, err := http.NewRequest(req.Method, req.URL, reqBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置请求头
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// 如果没有设置Content-Type且有数据，默认设置为JSON
	if req.Data != "" && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 解析响应头
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	response := map[string]interface{}{
		"status":  resp.StatusCode,
		"headers": headers,
		"body":    string(respBody),
	}

	c.JSON(http.StatusOK, response)
}

func listJSONFiles(c *gin.Context) {
	files, err := filepath.Glob(filepath.Join(getJSONFilesPath(currentProject), "*.json"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, filepath.Base(file))
	}

	c.JSON(http.StatusOK, fileNames)
}

func saveJSONFile(c *gin.Context) {
	var request struct {
		Filename string `json:"filename"`
		Content  string `json:"content"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证文件名安全性
	if strings.Contains(request.Filename, "..") || strings.Contains(request.Filename, "/") || strings.Contains(request.Filename, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "非法文件名"})
		return
	}

	// 验证JSON格式
	var jsonData interface{}
	if err := json.Unmarshal([]byte(request.Content), &jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON格式错误: " + err.Error()})
		return
	}

	// 保存文件
	filePath := filepath.Join(getJSONFilesPath(currentProject), request.Filename)
	if err := os.WriteFile(filePath, []byte(request.Content), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文件保存成功"})
}

func readJSONFile(c *gin.Context) {
	filename := c.Query("file")

	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件名不能为空"})
		return
	}

	// 验证文件名安全性
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "非法文件名"})
		return
	}

	// 读取文件
	filePath := filepath.Join(getJSONFilesPath(currentProject), filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败: " + err.Error()})
		return
	}

	// 解析JSON并返回
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件不是有效的JSON格式: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonData)
}

const globalConfigFileName = "config.json"

func saveConfig() error {
	server.mu.RLock()
	config := Config{
		IP:         server.IP,
		Port:       server.Port,
		Endpoints:  server.Endpoints,
		SendBlocks: server.SendBlocks,
	}
	server.mu.RUnlock()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	configPath := filepath.Join(getProjectPath(currentProject), "config.json")
	return os.WriteFile(configPath, data, 0644)
}

func loadConfig() error {
	configPath := filepath.Join(getProjectPath(currentProject), "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("项目配置文件不存在，使用默认配置")
			return nil
		}
		return err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	server.mu.Lock()
	server.IP = config.IP
	server.Port = config.Port
	server.Endpoints = config.Endpoints
	server.SendBlocks = config.SendBlocks
	server.mu.Unlock()

	log.Printf("项目 %s 配置文件加载成功", currentProject)
	return nil
}

func saveGlobalConfig() error {
	config := Config{
		IP:             server.IP,
		Port:           server.Port,
		CurrentProject: currentProject,
		Endpoints:      server.Endpoints,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(globalConfigFileName, data, 0644)
}

func loadGlobalConfig() error {
	data, err := os.ReadFile(globalConfigFileName)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("全局配置文件不存在，使用默认配置")
			currentProject = "default"
			return nil
		}
		return err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	if config.CurrentProject != "" {
		currentProject = config.CurrentProject
	}

	log.Printf("全局配置加载成功，当前项目: %s", currentProject)
	return nil
}

// 项目管理辅助函数
func getProjectPath(project string) string {
	return filepath.Join("projects", project)
}

func getJSONFilesPath(project string) string {
	return filepath.Join(getProjectPath(project), "json_files")
}

func getConfigPath(project string) string {
	return filepath.Join(getProjectPath(project), "config.json")
}

// 初始化项目结构
func initializeProjects() error {
	// 创建projects目录
	if err := os.MkdirAll("projects", 0755); err != nil {
		return err
	}

	// 检查default项目是否存在
	defaultPath := getProjectPath("default")
	if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
		// 创建default项目
		if err := os.MkdirAll(defaultPath, 0755); err != nil {
			return err
		}
		if err := os.MkdirAll(getJSONFilesPath("default"), 0755); err != nil {
			return err
		}

		// 迁移旧的json_files目录
		if _, err := os.Stat("json_files"); err == nil {
			log.Println("检测到旧的json_files目录，正在迁移到projects/default/...")
			files, _ := filepath.Glob("json_files/*.json")
			for _, file := range files {
				filename := filepath.Base(file)
				data, _ := os.ReadFile(file)
				os.WriteFile(filepath.Join(getJSONFilesPath("default"), filename), data, 0644)
			}
			log.Println("迁移完成")
		}

		log.Println("默认项目创建成功")
	}

	return nil
}

// API: 列出所有项目
func listProjects(c *gin.Context) {
	entries, err := os.ReadDir("projects")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var projects []ProjectInfo
	for _, entry := range entries {
		if entry.IsDir() {
			info, _ := entry.Info()
			projects = append(projects, ProjectInfo{
				Name:      entry.Name(),
				Path:      getProjectPath(entry.Name()),
				CreatedAt: info.ModTime().Format("2006-01-02 15:04:05"),
			})
		}
	}

	c.JSON(http.StatusOK, projects)
}

// API: 创建新项目
func createProject(c *gin.Context) {
	var request struct {
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证项目名
	if request.Name == "" || strings.Contains(request.Name, "..") || strings.Contains(request.Name, "/") || strings.Contains(request.Name, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "非法项目名"})
		return
	}

	projectPath := getProjectPath(request.Name)
	if _, err := os.Stat(projectPath); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目已存在"})
		return
	}

	// 创建项目目录
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建项目失败: " + err.Error()})
		return
	}

	// 创建json_files目录
	if err := os.MkdirAll(getJSONFilesPath(request.Name), 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建目录失败: " + err.Error()})
		return
	}

	// 创建默认配置文件
	defaultConfig := Config{
		IP:   "0.0.0.0",
		Port: "29800",
		Endpoints: []EndpointConfig{
			{Name: "测试接口1", Path: "/api/test1", ResponseFile: ""},
			{Name: "测试接口2", Path: "/api/test2", ResponseFile: ""},
			{Name: "测试接口3", Path: "/api/test3", ResponseFile: ""},
			{Name: "测试接口4", Path: "/api/test4", ResponseFile: ""},
		},
		SendBlocks: []SendBlock{
			{Name: "", URL: "", SendFile: "", Method: "POST", Headers: `{"content-type":"application/json"}`},
			{Name: "", URL: "", SendFile: "", Method: "POST", Headers: `{"content-type":"application/json"}`},
		},
	}

	data, _ := json.MarshalIndent(defaultConfig, "", "  ")
	os.WriteFile(getConfigPath(request.Name), data, 0644)

	c.JSON(http.StatusOK, gin.H{"message": "项目创建成功"})
}

// API: 切换项目
func switchProject(c *gin.Context) {
	var request struct {
		Project string `json:"project"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证项目是否存在
	projectPath := getProjectPath(request.Project)
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目不存在"})
		return
	}

	// 保存当前项目配置
	if err := saveConfig(); err != nil {
		log.Printf("保存当前项目配置失败: %v", err)
	}

	// 切换项目
	currentProject = request.Project

	// 加载新项目配置
	if err := loadConfig(); err != nil {
		log.Printf("加载新项目配置失败: %v", err)
	}

	// 保存全局配置
	if err := saveGlobalConfig(); err != nil {
		log.Printf("保存全局配置失败: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "项目切换成功"})
	broadcastToClients(map[string]interface{}{"type": "status_update", "data": server})
}