package main

import (
	"context"
	"fmt"
	"shortURL/internal/bootstrap"
	"shortURL/internal/model"
	"time"
)

// 测试接口
func serviceTest() {
	exp := 10
	appTest := &model.CreateURLRequest{
		LongURL:       "https://baidu.com",
		SelfShortCode: "selfTest",
		ExpireTime:    &exp,
	}
	app := bootstrap.SetUp()
	rep, err := app.Service.CreateURL(context.Background(), appTest)
	if err != nil {
		fmt.Println("测试失败", err)
	} else {
		fmt.Println("测试成功", rep)
	}

	time.Sleep(20 * time.Second)
}

func bloomFilterTest() {
	id := 178748247387439104
	app := bootstrap.SetUp()
	shortCode := app.Base62Generator.GenerateShortCode(int64(id))
	fmt.Println(shortCode)
	app.BloomFilter.AddBloomFilterElem([]byte(shortCode))
	if app.BloomFilter.IsExistData([]byte(shortCode)) {
		fmt.Println("元素可能存在")
	}
	if !app.BloomFilter.IsExistData([]byte("cvsda")) {
		fmt.Println("元素不存在")
	}
}

func main() {
	serviceTest()
}
