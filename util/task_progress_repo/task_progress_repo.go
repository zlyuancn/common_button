package task_progress_repo

import (
	"context"

	"github.com/zlyuancn/common_button/pb"
	"github.com/zlyuancn/common_button/util/task"
)

type Repo interface {
	// 批量查询任务进度
	MultiQueryTaskProgress(ctx context.Context, taskType pb.TaskType, tasks []task.Task) ([]int32, error)
}

var defRepo Repo = repoImpl{}

func SetRepo(repo Repo) {
	defRepo = repo
}
func GetRepo() Repo {
	return defRepo
}

type repoImpl struct{}

func (r repoImpl) MultiQueryTaskProgress(ctx context.Context, taskType pb.TaskType, tasks []task.Task) ([]int32, error) {
	//TODO implement me
	panic("implement me")
}
