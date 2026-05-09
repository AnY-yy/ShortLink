package base62

import "strings"

type ShortCodeGenerator interface {
	GenerateShortCode(snowFlakeID int64) string
}

type Base62 struct {
}

const characters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func NewShortGenerator() ShortCodeGenerator {
	return &Base62{}
}

func (b *Base62) GenerateShortCode(snowFlakeID int64) string {
	if snowFlakeID == 0 {
		return ""
	}
	// 根据雪花ID
	var result strings.Builder
	for snowFlakeID > 0 {
		reminder := snowFlakeID % 62
		result.WriteByte(characters[reminder])
		snowFlakeID /= 62
	}
	// 反转字符串
	code := b.reverseString(result.String())
	return code
}

func (b *Base62) reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
