package task

import "time"

type TaskRepo interface {
	CleanExpiredData()
	GetInterval() time.Duration
}

type Task interface {
	RunTask()
}

type TaskModel struct {
	taskRepo TaskRepo
}

func NewTask(taskRepo TaskRepo) Task {
	return &TaskModel{
		taskRepo: taskRepo,
	}
}

func (t *TaskModel) runTaskRepo() {
	// 定时器
	ticker := time.NewTicker(t.taskRepo.GetInterval())

	// 任务开始时直接执行一次
	t.taskRepo.CleanExpiredData()

	// 循环执行定时任务
	for range ticker.C {
		t.taskRepo.CleanExpiredData()
	}
}

// RunTask 启动全部定时任务
func (t *TaskModel) RunTask() {
	// 使用goroutine启动定时任务
	go t.runTaskRepo()
}
