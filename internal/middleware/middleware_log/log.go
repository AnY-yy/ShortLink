package middleware_log

import (
	"fmt"
	"math"
	"shortURL/internal/external/ipinfo"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
}

type MiddleWareLogger interface {
	RecordWebAction() gin.HandlerFunc
}

type IPInfoLoader interface {
	GetInfo(ip string) (*ipinfo.IPInfo, error)
}

type WebLogger struct {
	logger        Logger
	getClientInfo IPInfoLoader
}

func NewMiddleWareLogger(logger Logger, ipInfoLoader IPInfoLoader) MiddleWareLogger {
	return &WebLogger{
		logger:        logger,
		getClientInfo: ipInfoLoader,
	}
}

// 需要跳过日志记录的路径模式
var skipPaths = []string{
	"/favicon.ico",
	"/robots.txt",
	"/hybridaction",
	"/tracker",
	"/analytics",
	"/statistics",
}

// shouldSkipLog 判断是否应该跳过日志记录
func (w *WebLogger) shouldSkipLog(path string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// RecordWebAction 记录客户端web请求日志
func (w *WebLogger) RecordWebAction() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()

		stopTime := time.Since(startTime)
		t := fmt.Sprintf("%d ms", int(math.Ceil(float64(stopTime.Nanoseconds()/1000000.0))))

		// 跳过不需要记录的请求 避免记录静态资源、插件等的请求
		path := c.Request.URL.Path
		if w.shouldSkipLog(path) {
			return
		}

		ip := c.ClientIP()
		clientInfo, err := w.getClientInfo.GetInfo(ip)
		if err != nil {
			w.logger.Error(ip + "的信息获取失败")
		}

		// 日志输出
		w.logger.Info("web请求记录",
			zap.Any("client", clientInfo),                   // 客户端IP
			zap.String("method", c.Request.Method),          // 请求方法
			zap.String("path", path),                        // 请求路径
			zap.Int("status", c.Writer.Status()),            // 响应状态码
			zap.String("duration", t),                       // 响应时间
			zap.String("user_agent", c.Request.UserAgent()), // 用户代理
		)
	}
}
