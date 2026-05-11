package router

import (
	api_v2 "shortURL/internal/api-v2"
	"shortURL/internal/bootstrap"

	"github.com/gin-gonic/gin"
)

// InitRouter
// 传参新增App结构体指针 包含项目中的全部基础接口类型字段 用来注入依赖 集成结构体实例
func InitRouter(app *bootstrap.App) *gin.Engine {
	// 接口集成APIHandler结构体
	apiHandler := api_v2.NewAPIHandler(app.Logger, app.Service)
	// 集成中间件接口
	lg := app.MiddleWares.Logger
	ry := app.MiddleWares.Recover

	r := gin.New()
	r.Use(ry.Recovery(), lg.RecordWebAction())

	// 加载模板文件
	r.LoadHTMLFiles("./templates/index.tmpl")

	// 加载静态资源

	// 路由注册
	apiGroup := r.Group("/api/v2")
	{
		apiGroup.GET("/index", apiHandler.IndexHandler("", "", "", nil, nil))

		apiGroup.POST("/createurl", apiHandler.CreateURL)
	}

	r.GET("/:code", apiHandler.RedirectURL)

	return r
}
