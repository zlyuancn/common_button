package task

import (
	"context"
	"time"

	"github.com/zly-app/zapp/logger"
	"go.uber.org/zap"

	"github.com/zlyuancn/common_button/model"
	"github.com/zlyuancn/common_button/pb"
	"github.com/zlyuancn/common_button/util/hide_rule"
	"github.com/zlyuancn/common_button/util/task_period"
)

type Task interface {
	GetButton() *pb.Button
	GetUserTaskData() *model.UserTaskData

	// 判断是否需要持久化
	IsNeedPersistence(ctx context.Context) bool
	// 判断是否隐藏
	IsHide(ctx context.Context) (bool, error)
	// 判断是否需要查询任务进度
	IsNeedQueryTaskProgress(ctx context.Context) bool

	// 更新周期
	UpdatePeriod(ctx context.Context) error
	// 设置新进度
	SetNewProgress(ctx context.Context, progress int32)
}

type BaseTask struct {
	btn             *pb.Button
	td              *model.UserTaskData
	needPersistence bool
}

func (b *BaseTask) GetButton() *pb.Button {
	return b.btn
}

func (b *BaseTask) GetUserTaskData() *model.UserTaskData {
	return b.td
}

func (b *BaseTask) IsNeedPersistence(ctx context.Context) bool {
	return b.needPersistence
}

func (b *BaseTask) IsHide(ctx context.Context) (bool, error) {
	if !b.taskIsValid() {
		return true, nil
	}

	return hide_rule.CheckIsHide(ctx, b.btn, b.td)
}

func (b *BaseTask) IsNeedQueryTaskProgress(ctx context.Context) bool {
	// 未完成的任务需要查询任务进度
	return b.td.FinishStatus == pb.TaskFinishStatus_TASK_FINISH_STATUS_UNFINISHED
}

func (b *BaseTask) UpdatePeriod(ctx context.Context) error {
	mark, err := task_period.GenPeriodMark(ctx, b.btn, b.td)
	if err != nil {
		logger.Error(ctx, "UpdatePeriod call GenPeriodMark err", zap.Any("btn", b.btn), zap.Any("td", b.td), zap.Error(err))
		return err
	}

	if b.td.PeriodMark == mark {
		return nil
	}

	// 新的周期将数据重置
	b.td = &model.UserTaskData{
		PeriodMark:   mark,
		Progress:     0,
		FinishStatus: pb.TaskFinishStatus_TASK_FINISH_STATUS_UNFINISHED,
	}
	b.needPersistence = true
	b.fillTaskState()
	return nil
}

func (b *BaseTask) SetNewProgress(ctx context.Context, progress int32) {
	if progress == 0 {
		return
	}

	if b.td.FinishStatus != pb.TaskFinishStatus_TASK_FINISH_STATUS_UNFINISHED {
		return
	}

	if b.td.Progress != progress {
		b.needPersistence = true
	}

	b.td.Progress = progress
	if b.td.Progress >= b.btn.Task.TaskTarget {
		b.td.FinishStatus = pb.TaskFinishStatus_TASK_FINISH_STATUS_FINISHED
		b.needPersistence = true
	}

	b.fillTaskState()
}

// 填充任务状态
func (b *BaseTask) fillTaskState() {
	b.btn.TaskState = &pb.TaskState{
		TaskProgress: b.td.Progress,
		FinishStatus: b.td.FinishStatus,
	}
}

// 是否有效
func (b *BaseTask) taskIsValid() bool {
	t := int32(time.Now().Unix())
	if t >= b.btn.Task.StartTime && t < b.btn.Task.EndTime {
		return true
	}
	return false
}

var NewTask = func(btn *pb.Button, td *model.UserTaskData) Task {
	ret := &BaseTask{
		btn: btn,
		td:  td,
	}
	ret.fillTaskState()
	return ret
}
