package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// Server HTTP服务器
type Server struct {
	config     *Config
	apiHandler *APIHandler
	engine     *gin.Engine
	httpServer *http.Server
}

// NewServer 创建一个新的Server
func NewServer(config *Config) *Server {
	// 创建gin引擎
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	// 创建API处理器
	apiHandler := NewAPIHandler(config)

	// 创建HTTP服务器
	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	return &Server{
		config:     config,
		apiHandler: apiHandler,
		engine:     engine,
		httpServer: httpServer,
	}
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 健康检查
	s.engine.GET("/health", s.apiHandler.HandleHealth)

	// AI查询API
	s.engine.Any("/query", s.apiHandler.HandleQuery)
	s.engine.GET("/info", s.apiHandler.HandleInfo)

	// 首页
	s.engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "running",
			"name":   "AI答题本地服务",
		})
	})
}


// Start 启动服务器
func (s *Server) Start() error {
	// 设置路由
	s.setupRoutes()

	// 启动HTTP服务器
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	log.Printf("启动HTTP服务器: %s", addr)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP服务器启动失败: %v", err)
		}
	}()

	return nil
}

// Stop 停止服务器
func (s *Server) Stop() error {
	log.Println("正在停止HTTP服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("HTTP服务器关闭失败: %v", err)
	}

	log.Println("HTTP服务器已停止")
	return nil
}

// WaitForShutdown 等待关闭信号
func (s *Server) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("收到关闭信号，正在关闭服务器...")
	if err := s.Stop(); err != nil {
		log.Fatalf("服务器关闭失败: %v", err)
	}
}
