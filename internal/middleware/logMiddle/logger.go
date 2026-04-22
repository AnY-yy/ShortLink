package logMiddle

import (
	"fmt"
	"math"
	"shortURL/internal/bootstrap"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

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
func shouldSkipLog(path string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()

		stopTime := time.Since(startTime)
		spendTime := fmt.Sprintf("%d ms", int(math.Ceil(float64(stopTime.Nanoseconds()/1000000.0))))

		// 跳过不需要记录的请求 避免记录静态资源 插件等的请求
		path := c.Request.URL.Path
		if shouldSkipLog(path) {
			return
		}

		// 日志输出
		bootstrap.Application.Logger.Info("HTTP Request",
			zap.String("method", c.Request.Method),          // 请求方法
			zap.String("path", c.Request.URL.Path),          // 请求路径
			zap.Int("status", c.Writer.Status()),            // 响应状态码
			zap.String("duration", spendTime),               // 请求耗时
			zap.String("ip", c.ClientIP()),                  // 客户端IP
			zap.String("user_agent", c.Request.UserAgent()), // 用户代理
		)
	}
}
