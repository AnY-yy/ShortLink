package model

import "time"

type RedirectURLRequest struct {
	ShortURL string `json:"shorturl"`
}

type RedirectURLResponse struct {
	ShortURL string    `json:"shorturl"`
	LongURL  string    `json:"longurl"`
	ExpireAt time.Time `json:"expiretime"`
}
