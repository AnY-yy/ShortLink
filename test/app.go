package main

import (
	"context"
	"fmt"
	"shortURL/internal/bootstrap"
	"shortURL/internal/external/ipinfo"
	"shortURL/internal/model"
	"time"
)

// 服务层测试
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

// 布隆过滤器测试
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

// 第三方API - ipinfo.io测试
func ipInfoAPI() {
	// 云服务器ip: 154.9.253.151
	// ip := "154.9.253.151" // &{154.9.253.151 Hong Kong Hong Kong HK 22.2783,114.1747 AS979 NetLab Global}
	ip := "114.114.114.114" // &{114.114.114.114 Shanghai Shanghai CN 31.2222,121.4581 AS21859 Zenlayer Inc}

	// 配置接口注入
	app := bootstrap.SetUp()
	cfgProvider := app.Config

	// ipinfoop接口注入
	ipinfoop := ipinfo.NewIPInfoLoader(cfgProvider)

	err := ipinfoop.Ping()
	if err != nil {
		panic(err)
	}

	// 获取地址
	info, err := ipinfoop.GetInfo(ip)
	if err != nil {
		panic(err)
	}
	fmt.Println(info)
}

func main() {
	ipInfoAPI()
}
