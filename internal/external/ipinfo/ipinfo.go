package ipinfo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"shortURL/config"
)

type IPInfoLoader interface {
	Ping() error
	GetInfo(ip string) (*IPInfo, error)
}

type IPInfo struct {
	IP      string `json:"ip"`
	City    string `json:"city"`
	Region  string `json:"region"`
	Country string `json:"country"`
	Loc     string `json:"loc"`
	Org     string `json:"org"`
}

type IPInfoOP struct {
	Token string
}

func NewIPInfoLoader(cfgProvider config.ConfigProvider) IPInfoLoader {
	cfg := cfgProvider.GetConfig()

	return &IPInfoOP{
		Token: cfg.IPInfoCfg.Token,
	}
}

// Ping 测试URL是否可行
func (i *IPInfoOP) Ping() error {
	url := fmt.Sprintf("https://ipinfo.io/%s?token=%s", "8.8.8.8", i.Token) // 测试url

	rep, err := http.Get(url)
	if err != nil {
		return err
	}
	defer rep.Body.Close()
	return nil
}

// GetInfo 得到客户端信息
func (i *IPInfoOP) GetInfo(ip string) (*IPInfo, error) {
	rep, err := http.Get(fmt.Sprintf("https://ipinfo.io/%s?token=%s", ip, i.Token))
	if err != nil {
		return nil, err
	}
	defer rep.Body.Close()

	var info IPInfo
	json.NewDecoder(rep.Body).Decode(&info)
	return &info, nil
}
