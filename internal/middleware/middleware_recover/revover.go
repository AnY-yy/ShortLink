package middleware_recover

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

/*
捕获请求中发生的panic 防止整个服务挂掉
终止请求而不是挂掉服务
*/

type Logger interface {
	Error(msg string, fields ...zap.Field)
}

type MiddleWareRecover interface {
	Recovery() gin.HandlerFunc
}

type PanicRecover struct {
	logger Logger
}

func NewMiddleWareRecover(logger Logger) MiddleWareRecover {
	return &PanicRecover{
		logger: logger,
	}
}

func (p *PanicRecover) Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 捕获panic 日志记录
				p.logger.Error("请求发生panic",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("ip", c.ClientIP()),
				)

				// 给前端返回错误
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "服务器内部异常",
				})

				// 终止请求
				c.Abort()
			}
		}()

		c.Next()
	}
}
