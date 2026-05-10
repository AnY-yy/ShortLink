package api_v2

import (
	"context"
	"fmt"
	"net/http"
	"shortURL/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// Logger
// 日志接口 引用pkg/logger下的接口
type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
}

// ServiceCore
// 项目服务核心接口 包含短链创建与重定向服务
type ServiceCore interface {
	CreateURL(ctx context.Context, req *model.CreateURLRequest) (*model.CreateURLReponse, error)
}

type Handler interface {
	IndexHandler(longurl, shorturl, selfshorturl string, err error, expiretime *int) func(c *gin.Context)
	CreateURL(c *gin.Context)
	RedirectURL(c *gin.Context)
}

type APIHandler struct {
	logger      Logger
	serviceCore ServiceCore
	validator   *validator.Validate // 数据格式验证器
}

// NewAPIHandler
// 暴露结构体 router下通过结构体变量调用路由函数
func NewAPIHandler(logger Logger, serviceCore ServiceCore) Handler {
	return &APIHandler{
		logger:      logger,
		serviceCore: serviceCore,
		validator:   validator.New(),
	}
}

/*
templates/index.tmpl
可传入模板参数: map[string]interface{}
    "shorturl": "生成的短码",      // 可选
    "error": "错误信息",           // 可选
    "longurl": "原始长链接",       // 可选
    "selfshorturl": "自定义短码",  // 可选
    "expiretime": "过期时间",      // 可选
*/

// IndexHandler
// GET请求 /api/v2/index 返回首页
func (a *APIHandler) IndexHandler(longurl, shorturl, selfshorturl string, err error, expiretime *int) func(c *gin.Context) {
	return func(c *gin.Context) {
		// 判断指针类型是否为空
		data := make(map[string]interface{})
		data["shorturl"] = shorturl
		data["longurl"] = longurl
		data["selfshorturl"] = selfshorturl
		if expiretime != nil {
			data["expiretime"] = *expiretime
		} else {
			data[""] = nil
		}
		if err == nil {
			data["error"] = nil
		} else {
			data["error"] = err.Error()
		}
		// 渲染模板
		c.HTML(http.StatusOK, "index.tmpl", data)
	}
}

// CreateURL
// POST请求 /api/v2/createurl 创建短链
func (a *APIHandler) CreateURL(c *gin.Context) {
	a.logger.Info("/api/v2/createurl接收到POST请求, 函数/internal/api-v2/api.go/CreateURL开始执行")
	// 数据绑定
	a.logger.Info("开始绑定前端数据")
	var req model.CreateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Error("前端数据绑定失败", zap.Error(err))
		a.IndexHandler("", "", "", err, nil)(c)
		return
	}
	// 数据格式验证
	a.logger.Info("开始数据格式校验")
	if err := a.validator.Struct(req); err != nil {
		a.logger.Error("格式错误,数据格式校验错误", zap.Error(err))
		a.IndexHandler("", "", "", fmt.Errorf("数据格式错误: %w", err), nil)(c)
		return
	}
	// 创建短链服务
	a.logger.Info("开始执行创建短链服务,调用ServiceCore接口中的CreateURL方法")
	rep, err := a.serviceCore.CreateURL(c, &req)
	if err != nil {
		a.logger.Error("创建短链服务失败", zap.Error(err))
		a.IndexHandler("", "", "", fmt.Errorf("创建短链失败,%v", err), nil)(c)
		return
	}
	// 返回响应
	if rep != nil {
		a.logger.Info(req.LongURL+"的短链创建成功", zap.String(req.LongURL, rep.ShortCode))
		a.IndexHandler(req.LongURL, rep.ShortCode, req.SelfShortCode, nil, req.ExpireTime)(c)
		return
	}
}

// RedirectURL
// GET请求 /:code 重定向短链
func (a *APIHandler) RedirectURL(c *gin.Context) {
	// 获取URL参数 code
	code := c.Param("code")
	fmt.Println(code)
}
