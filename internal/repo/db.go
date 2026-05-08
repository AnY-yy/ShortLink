package repo

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	LongURLIsExist(ctx context.Context, longURL string) (bool, error)
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

	return false, nil
}
