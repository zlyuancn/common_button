package view

import (
	"context"
	"errors"
	"fmt"

	"github.com/mohae/deepcopy"
	"github.com/zly-app/zapp/logger"
	"go.uber.org/zap"

	"github.com/zlyuancn/common_button/dao"
	"github.com/zlyuancn/common_button/pb"
	"github.com/zlyuancn/common_button/util/task_repo"
)

type implCli struct {
	pb.UnimplementedCommonButtonServiceServer
}

func (impl implCli) GetButtonList(ctx context.Context, req *pb.GetButtonListReq) (*pb.GetButtonListRsp, error) {
	buttons, err := dao.GetButtonRepo().GetButtonsByModuleAndScene(ctx, req.ModuleId, req.SceneId)
	if err != nil {
		logger.Error(ctx, "GetButtonList call GetButtonsByModuleAndScene err", zap.Any("req", req), zap.Error(err))
		return nil, err
	}

	taskMM, err := task_repo.GetRepo().MultiGetTasksStatus(ctx, buttons)
	if err != nil {
		logger.Error(ctx, "GetButtonList call MultiGetTasksStatus err", zap.Any("req", req), zap.Error(err))
		return nil, err
	}

	// 过滤需要隐藏的按钮, 更新进度
	ret := make([]*pb.Button, 0, len(buttons))
	for _, btn := range buttons {
		if btn.Task == nil {
			ret = append(ret, btn)
			continue
		}

		t, ok := taskMM[btn.Task.TaskId]
		if !ok {
			// 理论上不会走到这个case
			logger.Error(ctx, "GetButtonList abnormal. not found task.", zap.Int32("taskID", btn.Task.TaskId), zap.Any("btn", btn))
			return nil, fmt.Errorf("GetButtonList abnormal. not found task. btnID=%d, taskID=%d", btn.ButtonId, btn.Task.TaskId)
		}

		// 隐藏
		hide, err := t.IsHide(ctx)
		if err != nil {
			logger.Error(ctx, "GetButtonList abnormal. check task IsHide err.", zap.Int32("taskID", btn.Task.TaskId), zap.Any("btn", btn), zap.Error(err))
			return nil, fmt.Errorf("GetButtonList abnormal. check task IsHide err. btnID=%d, taskID=%d, err=%v", btn.ButtonId, btn.Task.TaskId, err)
		}
		if hide {
			continue
		}

		btn = deepcopy.Copy(btn).(*pb.Button) // 修改数据必须深拷贝

		// 更新进度
		state := t.GetUserTaskData()
		btn.TaskState = &pb.TaskState{
			TaskProgress: state.Progress,
			FinishStatus: state.FinishStatus,
		}
		ret = append(ret, btn)
	}
	return &pb.GetButtonListRsp{Buttons: ret}, nil
}

func (implCli) ClickButton(ctx context.Context, req *pb.ClickButtonReq) (*pb.ClickButtonRsp, error) {
	btn, err := dao.GetButtonRepo().GetButtonByID(ctx, req.ButtonId)
	if err != nil {
		logger.Error(ctx, "ClickButton call GetButtonByID err", zap.Any("req", req), zap.Error(err))
		return nil, err
	}

	if btn.Task == nil {
		return &pb.ClickButtonRsp{TaskState: &pb.ClickTaskState{ButtonId: btn.ButtonId}}, nil
	}

	t, err := task_repo.GetRepo().ClickOneButton(ctx, btn)
	if err != nil {
		logger.Error(ctx, "ClickButton call ClickOneButton err", zap.Any("req", req), zap.Error(err))
		return nil, err
	}

	// 隐藏
	hide, err := t.IsHide(ctx)
	if err != nil {
		logger.Error(ctx, "ClickButton abnormal. check task IsHide err.", zap.Int32("taskID", btn.Task.TaskId), zap.Any("btn", btn), zap.Error(err))
		return nil, fmt.Errorf("ClickButton abnormal. check task IsHide err. btnID=%d, taskID=%d, err=%v", btn.ButtonId, btn.Task.TaskId, err)
	}
	ret := &pb.ClickTaskState{
		ButtonId:     btn.ButtonId,
		TaskId:       btn.Task.TaskId,
		FinishStatus: t.GetUserTaskData().FinishStatus,
		Prizes:       btn.Task.Prizes,
	}
	if hide {
		ret.FinishStatus = pb.TaskFinishStatus_TASK_FINISH_STATUS_Hide
	}
	return &pb.ClickButtonRsp{TaskState: ret}, nil
}

func (implCli) OneClickFinish(ctx context.Context, req *pb.OneClickFinishReq) (*pb.OneClickFinishRsp, error) {
	return nil, errors.New("未实现")
}

func NewButtonService() pb.CommonButtonServiceServer {
	return &implCli{}
}
