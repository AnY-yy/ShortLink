package repo

import (
	"fmt"
	"shortURL/internal/bootstrap"
	"shortURL/internal/model"

	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

// NewDB 返回数据库连接实例
func NewDB() *Repository {
	return &Repository{
		DB: bootstrap.Application.DB,
	}
}

// CreateURL 创建短链接数据
func (r *Repository) CreateURL(urlParams *model.URLParams) error {

	return nil
}

// IsExistLongURL 查询数据库 检查长链接是否存在
func (r *Repository) IsExistLongURL(longURL string) (bool, error) {
	var urlParams model.URLParams
	if err := r.DB.Where("longurl = ?", longURL).First(&urlParams).Error; err != nil {
		return false, err
	}
	if urlParams.Id == 0 {
		return false, fmt.Errorf("长链接不存在")
	}
	return true, nil
}

// IsExistShortURL 查询数据库 检查短链接是否存在
func (r *Repository) IsExistShortURL(shortURL string) (bool, error) {
	var urlParams model.URLParams
	if err := r.DB.Where("shorturl = ?", shortURL).First(&urlParams).Error; err != nil {
		return false, err
	}
	if urlParams.Id == 0 {
		return false, fmt.Errorf("短链接不存在")
	}
	return true, nil
}

// GetShortURL 查询数据库 根据长链接返回短链接
func (r *Repository) GetShortURL(LongURL string) (*model.CreateURLResponse, error) {
	var shortURL string
	if err := r.DB.Model(&model.URLParams{}).Select("shorturl").Where("longurl = ?", LongURL).Scan(&shortURL).Error; err != nil {
		return nil, err
	}
	return &model.CreateURLResponse{
		ShortURL: shortURL,
	}, nil
}
