package repo

import (
	"context"
	"errors"
	"shortURL/internal/model"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	LongURLIsExist(ctx context.Context, longURL string) (bool, error)
	GetShortCode(ctx context.Context, longURL string) (*model.CreateURLReponse, error)
	ShortURLIsExist(ctx context.Context, shortURL string) (bool, error)
	CreateURL(url *model.URLParams) error
	CleanExpiredData(ctx context.Context, now time.Time) (int64, error)
	RedirectURL(ctx context.Context, shortURL string) (*model.RedirectURLResponse, error)
	GetAllShortURL() ([]string, error)
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

// CleanExpiredData
// 清理过期数据
func (d *DBRepository) CleanExpiredData(ctx context.Context, now time.Time) (int64, error) {
	result := d.db.WithContext(ctx).Where("expireat IS NOT NULL AND expireat < ?", now).Delete(&model.URLParams{})

	return result.RowsAffected, result.Error
}

// RedirectURL
// 在数据库中查询短链对应长链信息
func (d *DBRepository) RedirectURL(ctx context.Context, shortURL string) (*model.RedirectURLResponse, error) {
	rep := &model.RedirectURLResponse{}
	err := d.db.WithContext(ctx).Where("shorturl = ?", shortURL).First(&rep).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return rep, nil
}

// GetAllShortURL
// 得到数据库中的全部短码数据
func (d *DBRepository) GetAllShortURL() ([]string, error) {
	var shortURLs []string

	err := d.db.Model(&model.BloomFilterInjection{}).Pluck("shorturl", &shortURLs).Error
	if err != nil {
		return nil, err
	}

	return shortURLs, nil
}
