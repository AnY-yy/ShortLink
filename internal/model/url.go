package model

import (
	"time"
)

// URLParams 短链接参数 用于数据存储
type URLParams struct {
	ID           int64     `gorm:"primaryKey;column:id"`
	LongURL      string    `gorm:"column:longurl"`
	ShortURL     string    `gorm:"column:shorturl"`
	SelfShortUrl string    `gorm:"column:selfshorturl"`
	IsCustom     bool      `gorm:"column:iscustom"`
	ExpireAt     time.Time `gorm:"column:expiretime"`
	CreatedAt    time.Time `gorm:"column:createdtime"`
}

func (URLParams) TableName() string {
	return "urls"
}

// CreateURLRequest
// 要求LongURL必写并且符合URL规范
// 自定义短码要求 如果传入了非空字符串 则要求长度在4-10之间 并且要求必须由数字0-9与大小写字母中的元素组成
// 时间设置为指针类型避免分不清0是传入的还是空值  要求时间如果是非空数据 则最小值为0 最大值为100 0代表永不过期
type CreateURLRequest struct {
	LongURL      string `json:"longurl" validate:"required,url"`
	SelfShortUrl string `json:"selfshorturl" validate:"omitempty,min=4,max=10,alphanum"`
	ExpireTime   *int   `json:"expiretime" validate:"omitempty,min=0,max=100"`
}
type CreateURLResponse struct {
	ShortURL string `json:"shorturl"`
}

type RedirectURLRequest struct {
	ShortURL string `json:"shorturl"`
}
type RedirectURLResponse struct {
	LongURL  string    `json:"longurl"`
	ShortURL string    `json:"shorturl"`
	ExpireAt time.Time `json:"expiretime"`
}
