package task_progress

import (
	"context"
	"fmt"

	"github.com/zlyuancn/common_button/pb"
	"github.com/zlyuancn/common_button/util/task"
)

type ProgressHandle interface {
	// 批量获取任务进度, 返回一个map[按钮id]进度值. 如果不返回某个按钮id的进度, 则其进度为0
	MultiQueryTaskProgress(ctx context.Context, tasks []task.Task) (map[int32]int32, error)
}

var progressHs = map[pb.TaskType]ProgressHandle{
	pb.TaskType_TASK_TYPE_UNKNOWN: noneProgress{},
	pb.TaskType_TASK_TYPE_JUMP:    noneProgress{},
	pb.TaskType_TASK_TYPE_CHECKIN: checkinProgress{},
}

func RegistryProgressHandle(t pb.TaskType, p ProgressHandle) {
	progressHs[t] = p
}

// 批量查询任务进度
func MultiQueryTaskProgress(ctx context.Context, taskType pb.TaskType, tasks []task.Task) (map[int32]int32, error) {
	h, ok := progressHs[taskType]
	if !ok {
		return nil, fmt.Errorf("Query TaskProgress TaskType is invalid. t=%d", int(taskType))
	}
	ret, err := h.MultiQueryTaskProgress(ctx, tasks)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

type noneProgress struct{}

func (n noneProgress) MultiQueryTaskProgress(ctx context.Context, tasks []task.Task) (map[int32]int32, error) {
	ret := make(map[int32]int32, 0)
	return ret, nil
}

type checkinProgress struct{}

func (n checkinProgress) MultiQueryTaskProgress(ctx context.Context, tasks []task.Task) (map[int32]int32, error) {
	ret := make(map[int32]int32, len(tasks))
	for i := 0; i < len(ret); i++ {
		ret[tasks[i].GetButton().ButtonId] = tasks[i].GetButton().Task.TaskTarget
	}
	return ret, nil
}
