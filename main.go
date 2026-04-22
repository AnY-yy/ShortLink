package main

import (
	"shortURL/internal/bootstrap"
	"shortURL/internal/router"
)

var err error

func main() {
	// 初始化组件
	bootstrap.Setup()

	// 启动服务
	router.InitRouter().Run(":8080")
}
