package model

// CreateURLRequest
// 创建短链接请求参数结构体
type CreateURLRequest struct {
	LongURL       string `json:"longurl" validate:"required,url"`
	SelfShortCode string `json:"selfshorturl" validate:"omitempty,min=4,max=10,alphanum"`
	ExpireTime    *int   `json:"expiretime" validate:"omitempty,min=0,max=100"`
}

// CreateURLReponse
// 创建短链接响应参数结构体
type CreateURLReponse struct {
	ShortCode string `json:"shorturl"`
}
