package api

import (
	"context"
	"fmt"
	"net/http"
	"shortURL/internal/bootstrap"
	"shortURL/internal/model"

	"github.com/gin-gonic/gin"
	validator2 "github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type URLService interface {
	CreateURL(ctx context.Context, req *model.CreateURLRequest) (*model.CreateURLResponse, error)
	RedirectURL(ctx context.Context, req *model.RedirectURLRequest) (*model.RedirectURLResponse, error)
}
type URLHandler struct {
	urlService URLService           // 包含短链的主要服务接口
	validator  *validator2.Validate // 校验器
}

// NewHandler 将api函数接口暴露
func NewHandler(service URLService) *URLHandler { // service 是业务逻辑层的接口
	return &URLHandler{
		urlService: service,
		validator:  validator2.New(),
	}
}

// IndexHandler
// GET请求 /api/v1/index 渲染首页界面
func (uh *URLHandler) IndexHandler(rep *model.CreateURLResponse) func(c *gin.Context) {
	return func(c *gin.Context) {
		var data gin.H
		if rep != nil {
			data = gin.H{
				"shorturl": rep.ShortURL,
			}
		} else {
			data = nil
		}
		c.HTML(http.StatusOK, "index.tmpl", data)
	}
}

// CreateURL 创建长链对应的唯一短链
// 调用真实业务逻辑层的CreateURL方法  这个方法的接口应该暴露给外部调用
// POST请求 /api/v1/createurl
func (uh *URLHandler) CreateURL(c *gin.Context) {
	// 数据获取
	var req model.CreateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 数据校验
	if err := uh.validator.Struct(req); err != nil {
		bootstrap.Application.Logger.Error("输入数据格式错误", zap.Error(err))
		return
	}
	fmt.Println(req)

	// 业务逻辑层调用 传入数据格式正确的req
	rep, err := uh.urlService.CreateURL(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, rep)
	// 渲染首页界面 将创建的短链返回给客户端
	if rep != nil {
		uh.IndexHandler(rep)(c)
	}
}
