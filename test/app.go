package main

import (
	"fmt"
	"shortURL/internal/bootstrap"
)

func main() {
	bootstrap.Setup()
	// 测试雪花ID生成
	for i := 0; i < 100; i++ {
		id := bootstrap.Application.SnowFlake.GenerateSnowFlakeID()
		fmt.Printf("id: %d\n", id)
	}
}
