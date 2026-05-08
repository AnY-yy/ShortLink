package model

import "time"

type URLParams struct {
	ID           int64     `gorm:"primaryKey;column:id"`
	LongURL      string    `gorm:"column:longurl"`
	ShortURL     string    `gorm:"column:shorturl"`
	SelfShortUrl string    `gorm:"column:selfshorturl"`
	IsCustom     bool      `gorm:"column:iscustom"`
	ExpireAt     time.Time `gorm:"column:expireat"`
	CreatedAt    time.Time `gorm:"column:createat"`
}

func (URLParams) TableName() string {
	return "v2"
}
