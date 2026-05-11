package repotask

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type Logger interface {
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
}

type Repo interface {
	CleanExpiredData(ctx context.Context, now time.Time) (int64, error)
}

type TaskRepo interface {
	CleanExpiredData()
	GetInterval() time.Duration
}

type TaskRepoModel struct {
	logger   Logger
	taskName string
	taskCore Repo
	interval time.Duration // 定时任务的执行间隔
}

func NewTaskRepo(logger Logger, taskName string, taskCore Repo, interval time.Duration) TaskRepo {
	return &TaskRepoModel{
		logger:   logger,
		taskName: taskName,
		taskCore: taskCore,
		interval: interval,
	}
}

// CleanExpiredData 定时任务: 查表、删除过期数据
func (t *TaskRepoModel) CleanExpiredData() {
	t.logger.Info("开始执行定时任务:", zap.String("name", t.taskName), zap.Duration("interval", t.interval))

	// 设置上下文超时 10s 避免卡死
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	defer cancel()

	// 执行核心逻辑
	now := time.Now()
	rows, err := t.taskCore.CleanExpiredData(ctx, now)
	if err != nil {
		t.logger.Warn("定时任务:"+t.taskName+"执行失败", zap.Error(err))
		return
	}
	if rows == 0 {
		t.logger.Info("暂无过期数据")
		return
	}
	t.logger.Info("删除过期数据成功,删除数量:", zap.Int64("rows", rows))
}

// GetInterval 得到定时任务间隔
func (t *TaskRepoModel) GetInterval() time.Duration {
	return t.interval
}
