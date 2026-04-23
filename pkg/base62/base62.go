package base62

import "strings"

type ShortCodeGenerator struct {
}

const characters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" // a-z A-Z 0-9 62个字符

// NewShortCodeGenerator 创建短链生成器 暴露给外部调用
func NewShortCodeGenerator() *ShortCodeGenerator {
	return &ShortCodeGenerator{}
}

// GenerateShortCode 生成短码
func (s *ShortCodeGenerator) GenerateShortCode(snowflakeID int64) string {
	if snowflakeID == 0 { // 防止非法传入0值
		return ""
	}
	// 将雪花ID转换为62进制字符串
	// 获得的62进制结果是低位在前 高位在后 所以需要反转字符串
	// 比如125 第一次取余得到1 第二次取余得到2 拼接得到的字符串是12 但正确顺序是21 所以需要反转字符串
	// 拼接字符优化: 运算符合+ -> strings.Builder
	var result strings.Builder
	for snowflakeID > 0 {
		reminder := snowflakeID % 62           // 得到小于62的余数 去得到对应的字符串
		result.WriteByte(characters[reminder]) // 先获得低位字符
		snowflakeID /= 62                      // 再进行高位字符的计算
	}
	// 反转字符串
	code := ReverseString(result.String())
	return code
}

// ReverseString  反转字符串
func ReverseString(s string) string {
	// go中string是不可变的 所以需要将string转换为rune切片
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
