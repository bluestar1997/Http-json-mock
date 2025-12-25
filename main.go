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
	RequestLogs       []RequestLog           `json:"request_logs"`
	mu                sync.RWMutex
	engine            *gin.Engine
	upgrader          websocket.Upgrader
	clients           map[*websocket.Conn]bool
	clientsMu         sync.RWMutex
}

type EndpointConfig struct {
	Path         string `json:"path"`
	ResponseFile string `json:"response_file"`
	IsActive     bool   `json:"is_active"`
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
	IP        string           `json:"ip"`
	Port      string           `json:"port"`
	Endpoints []EndpointConfig `json:"endpoints"`
}

var server *Server

func init() {
	server = &Server{
		IP:        "192.168.1.100",
		Port:      "29800",
		IsRunning: false,
		Endpoints: []EndpointConfig{
			{Path: "/api/audioTask/getAuditTaskResult", ResponseFile: "", IsActive: true},
			{Path: "/api/test1", ResponseFile: "", IsActive: true},
			{Path: "/api/test2", ResponseFile: "", IsActive: true},
			{Path: "/api/test3", ResponseFile: "", IsActive: true},
			{Path: "/api/test4", ResponseFile: "", IsActive: true},
		},
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
	// 创建必要的目录和文件
	os.MkdirAll("templates", 0755)
	os.MkdirAll("static", 0755)
	os.MkdirAll("json_files", 0755)

	// 加载配置文件
	if err := loadConfig(); err != nil {
		log.Printf("加载配置文件失败: %v", err)
	}

	createHTMLTemplate()
	createCSSFile()
	createJSFile()
	createSampleJSONFiles()

	r := gin.Default()

	// 静态文件服务
	r.Static("/static", "./static")
	r.Static("/json_files", "./json_files")
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

	c.JSON(http.StatusOK, server)
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
		if endpoint.IsActive {
			path := endpoint.Path
			responseFile := endpoint.ResponseFile

			server.engine.Any(path, func(c *gin.Context) {
				handleDynamicEndpoint(c, path, responseFile)
			})
		}
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
		IP        string           `json:"ip"`
		Port      string           `json:"port"`
		Endpoints []EndpointConfig `json:"endpoints"`
	}

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	server.mu.Lock()
	server.IP = config.IP
	server.Port = config.Port
	server.Endpoints = config.Endpoints
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
		data, err := os.ReadFile(filepath.Join("json_files", responseFile))
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
	files, err := filepath.Glob("json_files/*.json")
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
	filePath := filepath.Join("json_files", request.Filename)
	if err := os.WriteFile(filePath, []byte(request.Content), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文件保存成功"})
}

const configFileName = "config.json"

func saveConfig() error {
	server.mu.RLock()
	config := Config{
		IP:        server.IP,
		Port:      server.Port,
		Endpoints: server.Endpoints,
	}
	server.mu.RUnlock()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFileName, data, 0644)
}

func loadConfig() error {
	data, err := os.ReadFile(configFileName)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("配置文件不存在，使用默认配置")
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
	server.mu.Unlock()

	log.Println("配置文件加载成功")
	return nil
}