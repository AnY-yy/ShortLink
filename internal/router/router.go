package router

import (
	"shortURL/internal/api"
	"shortURL/internal/cache"
	"shortURL/internal/middleware/logMiddle"
	"shortURL/internal/repo"
	"shortURL/internal/service"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(logMiddle.Logger())

	// 解析模板
	r.LoadHTMLFiles("./templates/index.tmpl")

	// 服务接口注册
	// repository结构体实例 实现了Repository接口  用于数据库操作
	repository := repo.NewDB()
	// cache结构体实例 实现了Cache接口  用于缓存操作
	cacheRDB := cache.NewCache()
	// 得到服务层的结构体实例 用于业务逻辑层调用 核心服务接口
	urlService := service.NewURLService(repository, cacheRDB)
	// 得到api层的结构体实例 用于处理http请求 包含上面全部的服务接口 可灵活调用各种服务接口
	urlHandler := api.NewHandler(urlService)

	// 路由注册
	r.GET("/", urlHandler.IndexHandler(nil))

	apiV1 := r.Group("/api/v1")
	{
		apiV1.GET("/index", urlHandler.IndexHandler(nil))
		apiV1.POST("/createurl", urlHandler.CreateURL)
	}

	r.GET("/:code", urlHandler.RedirectURL)

	return r
}
