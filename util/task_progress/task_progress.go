package task_progress

import (
	"context"
	"fmt"

	"github.com/zlyuancn/common_button/pb"
	"github.com/zlyuancn/common_button/util/task"
)

type ProgressHandle interface {
	MultiQueryTaskProgress(ctx context.Context, tasks []task.Task) ([]int32, error)
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
func MultiQueryTaskProgress(ctx context.Context, taskType pb.TaskType, tasks []task.Task) ([]int32, error) {
	h, ok := progressHs[taskType]
	if !ok {
		return nil, fmt.Errorf("Query TaskProgress TaskType is invalid. t=%d", int(taskType))
	}
	ret, err := h.MultiQueryTaskProgress(ctx, tasks)
	if err != nil {
		return nil, err
	}
	if len(ret) != len(tasks) {
		return nil, fmt.Errorf("Query TaskProgress TaskType=%d return data length=%d not match in tasks length=%d", int(taskType), len(ret), len(tasks))
	}
	return ret, nil
}

type noneProgress struct{}

func (n noneProgress) MultiQueryTaskProgress(ctx context.Context, tasks []task.Task) ([]int32, error) {
	ret := make([]int32, 0)
	return ret, nil
}

type checkinProgress struct{}

func (n checkinProgress) MultiQueryTaskProgress(ctx context.Context, tasks []task.Task) ([]int32, error) {
	ret := make([]int32, len(tasks))
	for i := 0; i < len(ret); i++ {
		ret[i] = tasks[i].GetButton().Task.TaskTarget
	}
	return ret, nil
}
