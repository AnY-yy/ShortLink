package repo

import (
	"context"
	"errors"
	"shortURL/internal/model"

	"gorm.io/gorm"
)

type Repository interface {
	LongURLIsExist(ctx context.Context, longURL string) (bool, error)
	GetShortCode(ctx context.Context, longURL string) (*model.CreateURLReponse, error)
	ShortURLIsExist(ctx context.Context, shortURL string) (bool, error)
	CreateURL(url *model.URLParams) error
}

type DBRepository struct {
	db *gorm.DB
}

// NewRepository 暴露实例 实现Repository接口
func NewRepository(db *gorm.DB) Repository {
	return &DBRepository{
		db: db,
	}
}

// LongURLIsExist 查询长链接是否生成了短链
func (d *DBRepository) LongURLIsExist(ctx context.Context, longURL string) (bool, error) {
	var url = model.URLParams{}
	if err := d.db.WithContext(ctx).Where("longurl = ?", longURL).First(&url).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { // 表中不存在该记录
			return false, nil
		}
		return false, err // 查询出现错误
	}
	if url.ID == 0 {
		return false, nil
	}
	return true, nil
}

// GetShortCode
// 根据长链接查询短链 返回短链响应体
func (d *DBRepository) GetShortCode(ctx context.Context, longURL string) (*model.CreateURLReponse, error) {
	var url = model.URLParams{}
	err := d.db.WithContext(ctx).Where("longurl = ?", longURL).First(&url).Error
	if err != nil {
		return nil, err
	}
	return &model.CreateURLReponse{
		ShortCode: url.ShortURL,
	}, nil
}

// ShortURLIsExist
// 查询短链是否存在
func (d *DBRepository) ShortURLIsExist(ctx context.Context, shortURL string) (bool, error) {
	var url = model.URLParams{}
	err := d.db.WithContext(ctx).Where("longurl = ?", shortURL).First(&url).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { // 表中不存在的err
			return false, nil
		}
		return false, err // 查询出现错误
	}
	if url.ID == 0 {
		return false, nil
	}
	return true, nil
}

// CreateURL
// 将短链参数写入到数据库中
func (d *DBRepository) CreateURL(url *model.URLParams) error {
	err := d.db.Create(&url).Error
	if err != nil {
		return err
	}
	return nil
}
