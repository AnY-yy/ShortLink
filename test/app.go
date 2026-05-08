package main

import (
	"fmt"
	"shortURL/internal/bootstrap"
)

// 测试接口
func main() {
	app := bootstrap.SetUp()
	fmt.Println(app)
}
