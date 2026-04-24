package main

import (
	"fmt"
	"shortURL/internal/bootstrap"
	"shortURL/pkg/base62"
)

func main() {
	bootstrap.Setup()
	// 测试雪花ID+短码生成
	for i := 0; i < 100; i++ {
		go func(i int) {
			id := bootstrap.Application.SnowFlake.GenerateSnowFlakeID()
			shortCode := base62.NewShortCodeGenerator().GenerateShortCode(id)
			fmt.Printf("%d: %s\n", id, shortCode)
		}(i)
	}
}
