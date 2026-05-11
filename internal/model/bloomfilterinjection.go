package model

type BloomFilterInjection struct {
	shortURL string `gorm:"shorturl"`
}

func (BloomFilterInjection) TableName() string {
	return "v2"
}
