package view

import (
	"context"
	"errors"
	"fmt"

	"github.com/mohae/deepcopy"
	"github.com/zly-app/zapp/logger"
	"go.uber.org/zap"

	"github.com/zlyuancn/common_button/dao"
	"github.com/zlyuancn/common_button/dao/user_task_data_repo"
	"github.com/zlyuancn/common_button/pb"
	"github.com/zlyuancn/common_button/util/task_repo"
	"github.com/zlyuancn/common_button/util/user_repo"
)

type implCli struct {
	pb.UnimplementedCommonButtonServiceServer
}

func (impl implCli) GetButtonList(ctx context.Context, req *pb.GetButtonListReq) (*pb.GetButtonListRsp, error) {
	buttons, err := dao.GetButtonRepo().GetButtonsByModuleAndScene(ctx, req.ModuleId, req.SceneId...)
	if err != nil {
		logger.Error(ctx, "GetButtonList call GetButtonsByModuleAndScene err", zap.Any("req", req), zap.Error(err))
		return nil, err
	}

	// 分析有哪些带任务的按钮
	taskButtons := impl.matchTaskButton(ctx, buttons)
	if len(taskButtons) == 0 {
		return &pb.GetButtonListRsp{}, nil
	}

	// 解析uid
	uid, err := user_repo.GetRepo().ParseUID(ctx)
	if err != nil {
		logger.Error(ctx, "GetButtonList call user_repo.GetRepo().ParseUID err", zap.Error(err))
		return nil, err
	}

	// 加锁
	unlock, ok, err := user_task_data_repo.GetRepo().LockUser(ctx, uid)
	if err != nil {
		logger.Error(ctx, "GetButtonList call user_task_data_repo.GetRepo().LockUser err", zap.Error(err))
		return nil, err
	}
	if !ok {
		err = errors.New("get lock err")
		logger.Error(ctx, "GetButtonList call user_task_data_repo.GetRepo().LockUser err", zap.Error(err))
		return nil, err
	}
	defer unlock(ctx)

	// 获取任务状态
	taskMM, err := task_repo.GetRepo().MultiGetTasksStatus(ctx, uid, taskButtons)
	if err != nil {
		logger.Error(ctx, "GetButtonList call MultiGetTasksStatus err", zap.Any("req", req), zap.Error(err))
		return nil, err
	}
	unlock(ctx) // 立即解锁

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

// 找出有任务的按钮
func (implCli) matchTaskButton(ctx context.Context, buttons []*pb.Button) []*pb.Button {
	taskButtons := make([]*pb.Button, 0)
	for _, b := range buttons {
		if b.Task != nil {
			taskButtons = append(taskButtons, b)
		}
	}
	return taskButtons
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

	// 解析uid
	uid, err := user_repo.GetRepo().ParseUID(ctx)
	if err != nil {
		logger.Error(ctx, "ClickOneButton call user_repo.GetRepo().ParseUID err", zap.Error(err))
		return nil, err
	}

	// 加锁
	unlock, ok, err := user_task_data_repo.GetRepo().LockUser(ctx, uid)
	if err != nil {
		logger.Error(ctx, "ClickOneButton call user_task_data_repo.GetRepo().LockUser err", zap.Error(err))
		return nil, err
	}
	if !ok {
		err = errors.New("get lock err")
		logger.Error(ctx, "ClickOneButton call user_task_data_repo.GetRepo().LockUser err", zap.Error(err))
		return nil, err
	}
	defer unlock(ctx)

	// 点击按钮
	t, err := task_repo.GetRepo().ClickOneButton(ctx, uid, btn)
	if err != nil {
		logger.Error(ctx, "ClickButton call ClickOneButton err", zap.Any("req", req), zap.Error(err))
		return nil, err
	}
	unlock(ctx) // 立即解锁

	ret := &pb.ClickTaskState{
		ButtonId:     btn.ButtonId,
		TaskId:       btn.Task.TaskId,
		FinishStatus: t.GetUserTaskData().FinishStatus,
		Prizes:       btn.Task.Prizes,
	}
	hide, err := t.IsHide(ctx)
	if err != nil {
		logger.Error(ctx, "ClickButton abnormal. check task IsHide err.", zap.Int32("taskID", btn.Task.TaskId), zap.Any("btn", btn), zap.Error(err))
		return nil, fmt.Errorf("ClickButton abnormal. check task IsHide err. btnID=%d, taskID=%d, err=%v", btn.ButtonId, btn.Task.TaskId, err)
	}
	if hide {
		ret.FinishStatus = pb.TaskFinishStatus_TASK_FINISH_STATUS_Hide
	}
	return &pb.ClickButtonRsp{TaskState: ret}, nil
}

func (impl implCli) OneClickFinish(ctx context.Context, req *pb.OneClickFinishReq) (*pb.OneClickFinishRsp, error) {
	buttons, err := dao.GetButtonRepo().GetButtonsByModuleAndScene(ctx, req.ModuleId, req.SceneId)
	if err != nil {
		logger.Error(ctx, "OneClickFinish call GetButtonsByModuleAndScene err", zap.Any("req", req), zap.Error(err))
		return nil, err
	}

	// 分析有哪些带任务的按钮
	taskButtons := impl.matchTaskButton(ctx, buttons)
	if len(taskButtons) == 0 {
		return &pb.OneClickFinishRsp{}, nil
	}

	// 解析uid
	uid, err := user_repo.GetRepo().ParseUID(ctx)
	if err != nil {
		logger.Error(ctx, "OneClickFinish call user_repo.GetRepo().ParseUID err", zap.Error(err))
		return nil, err
	}

	// 加锁
	unlock, ok, err := user_task_data_repo.GetRepo().LockUser(ctx, uid)
	if err != nil {
		logger.Error(ctx, "OneClickFinish call user_task_data_repo.GetRepo().LockUser err", zap.Error(err))
		return nil, err
	}
	if !ok {
		err = errors.New("get lock err")
		logger.Error(ctx, "OneClickFinish call user_task_data_repo.GetRepo().LockUser err", zap.Error(err))
		return nil, err
	}
	defer unlock(ctx)

	// 批量领取
	taskMM, err := task_repo.GetRepo().MultiFinishTasks(ctx, uid, buttons)
	if err != nil {
		logger.Error(ctx, "OneClickFinish call MultiFinishTasks err", zap.Any("req", req), zap.Error(err))
		return nil, err
	}
	unlock(ctx) // 立即解锁

	ret := make([]*pb.ClickTaskState, 0, len(taskMM))
	for _, btn := range taskButtons {
		t, ok := taskMM[btn.Task.TaskId]
		if !ok {
			continue
		}
		state := &pb.ClickTaskState{
			ButtonId:     btn.ButtonId,
			TaskId:       btn.Task.TaskId,
			FinishStatus: t.GetUserTaskData().FinishStatus,
			Prizes:       btn.Task.Prizes,
		}
		hide, err := t.IsHide(ctx)
		if err != nil {
			logger.Error(ctx, "OneClickFinish abnormal. check task IsHide err.", zap.Int32("taskID", btn.Task.TaskId), zap.Any("btn", btn), zap.Error(err))
			return nil, fmt.Errorf("OneClickFinish abnormal. check task IsHide err. btnID=%d, taskID=%d, err=%v", btn.ButtonId, btn.Task.TaskId, err)
		}
		if hide {
			state.FinishStatus = pb.TaskFinishStatus_TASK_FINISH_STATUS_Hide
		}
		ret = append(ret, state)
	}

	return &pb.OneClickFinishRsp{TaskState: ret}, nil
}

func NewButtonService() pb.CommonButtonServiceServer {
	return &implCli{}
}
