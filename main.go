package main

import (
	"shortURL/internal/bootstrap"
	"shortURL/internal/router"
)

func main() {
	app := bootstrap.SetUp()

	r := router.InitRouter(app)

	err := r.Run(":8080")
	if err != nil {
		app.Logger.Panic("应用启动失败!")
	}
}
