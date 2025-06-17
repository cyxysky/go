package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SimpleTestHandler 简单测试处理器，不需要数据库
type SimpleTestHandler struct{}

func NewSimpleTestHandler() *SimpleTestHandler {
	return &SimpleTestHandler{}
}

// GetTime 获取当前时间
func (h *SimpleTestHandler) GetTime(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"current_time": time.Now().Format("2006-01-02 15:04:05"),
		"timestamp":    time.Now().Unix(),
		"message":      "服务器时间获取成功",
	})
}

// Echo 回声测试
func (h *SimpleTestHandler) Echo(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"echo":       req,
		"received_at": time.Now().Format("2006-01-02 15:04:05"),
		"message":    "回声测试成功",
	})
}

// GetSystemInfo 获取系统信息
func (h *SimpleTestHandler) GetSystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"golang_version": "1.21+",
		"gin_version":    "1.9.1",
		"database":       "PostgreSQL (未连接)",
		"cache":          "Redis (未连接)",
		"status":         "服务正常运行",
		"endpoints": []string{
			"GET /api/v1/health",
			"GET /api/v1/test/time",
			"POST /api/v1/test/echo",
			"GET /api/v1/test/info",
		},
	})
} 