package task

import (
	"context"
	"time"

	"github.com/zly-app/zapp/logger"
	"go.uber.org/zap"

	"github.com/zlyuancn/common_button/dao/prize_repo"
	"github.com/zlyuancn/common_button/model"
	"github.com/zlyuancn/common_button/pb"
	"github.com/zlyuancn/common_button/util/hide_rule"
	"github.com/zlyuancn/common_button/util/task_period"
)

type Task interface {
	GetButton() *pb.Button
	GetUserTaskData() *model.UserTaskData

	// 判断是否需要持久化, 会立即清除需要持久化标记
	IsNeedPersistence(ctx context.Context) bool
	// 判断是否隐藏
	IsHide(ctx context.Context) (bool, error)
	// 判断是否需要查询任务进度
	IsNeedQueryTaskProgress(ctx context.Context) bool

	// 更新周期
	UpdatePeriod(ctx context.Context) error
	// 设置新进度
	SetNewProgress(ctx context.Context, progress int32)

	// 点击按钮扭转状态
	ClickButton(ctx context.Context) error
}

type BaseTask struct {
	uid             string
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
	ret := b.needPersistence
	b.needPersistence = false
	return ret
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
}

func (b *BaseTask) ClickButton(ctx context.Context) error {
	// 特殊处理. 对跳转类型增加完成度
	switch b.btn.Task.TaskType {
	case pb.TaskType_TASK_TYPE_JUMP:
		progress := b.td.Progress + 1
		b.SetNewProgress(ctx, progress)
	}

	// 发货
	if b.td.FinishStatus == pb.TaskFinishStatus_TASK_FINISH_STATUS_FINISHED && len(b.btn.Task.Prizes) > 0 {
		err := prize_repo.GetRepo().SendPrize(ctx, b.uid, b.btn)
		if err != nil {
			logger.Error(ctx, "ClickButton call prize_repo.GetRepo().SendPrize err", zap.String("uid", b.uid), zap.Int32("buttonID", b.btn.ButtonId), zap.Error(err))
			return err
		}

		b.td.FinishStatus = pb.TaskFinishStatus_TASK_FINISH_STATUS_RECEIVED
		b.needPersistence = true
	}

	return nil
}

// 是否有效
func (b *BaseTask) taskIsValid() bool {
	t := int32(time.Now().Unix())
	if t >= b.btn.Task.StartTime && t < b.btn.Task.EndTime {
		return true
	}
	return false
}

var NewTask = func(uid string, btn *pb.Button, td *model.UserTaskData) Task {
	ret := &BaseTask{
		uid: uid,
		btn: btn,
		td:  td,
	}
	return ret
}
