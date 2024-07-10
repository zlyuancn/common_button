package task

import (
	"context"

	"github.com/zlyuancn/common_button/model"
	"github.com/zlyuancn/common_button/pb"
)

type Task interface {
	GetButton() *pb.Button
	GetUserTaskData() *model.UserTaskData

	// 判断是否需要持久化
	IsNeedPersistence(ctx context.Context) bool
	// 判断是否隐藏
	IsHide(ctx context.Context) bool
	// 判断是否需要查询任务进度
	IsNeedQueryTaskProgress(ctx context.Context) bool

	// 更新周期
	UpdatePeriod(ctx context.Context) error
	// 设置新进度
	SetNewProgress(ctx context.Context, progress int32)
}

type BaseTask struct {
	btn *pb.Button
	td  *model.UserTaskData
}

func (b *BaseTask) GetButton() *pb.Button {
	return b.btn
}

func (b *BaseTask) GetUserTaskData() *model.UserTaskData {
	return b.td
}

func (b *BaseTask) IsNeedPersistence(ctx context.Context) bool {
	//TODO implement me
	panic("implement me")
}

func (b *BaseTask) IsHide(ctx context.Context) bool {
	//TODO implement me
	panic("implement me")
}

func (b *BaseTask) IsNeedQueryTaskProgress(ctx context.Context) bool {
	//TODO implement me
	panic("implement me")
}

func (b *BaseTask) UpdatePeriod(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (b *BaseTask) SetNewProgress(ctx context.Context, progress int32) {
	//TODO implement me
	panic("implement me")
}

var NewTask = func(btn *pb.Button, td *model.UserTaskData) Task {
	ret := &BaseTask{
		btn: btn,
		td:  td,
	}
	return ret
}
