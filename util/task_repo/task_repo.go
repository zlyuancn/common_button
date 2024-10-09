package task_repo

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/logger"
	"go.uber.org/zap"

	"github.com/zlyuancn/common_button/dao/user_task_data_repo"
	"github.com/zlyuancn/common_button/model"
	"github.com/zlyuancn/common_button/pb"
	"github.com/zlyuancn/common_button/util/task"
	"github.com/zlyuancn/common_button/util/task_progress"
)

type Repo interface {
	// 批量获取任务状态
	MultiGetTasksStatus(ctx context.Context, uid string, buttons []*pb.Button) (map[int32]task.Task, error)
	// 点击一个按钮扭转状态
	ClickOneButton(ctx context.Context, uid string, button *pb.Button) (task.Task, error)
	// 批量领取任务奖励
	MultiFinishTasks(ctx context.Context, uid string, buttons []*pb.Button) (map[int32]task.Task, error)
}

var defRepo Repo = repoImpl{}

func SetRepo(repo Repo) {
	defRepo = repo
}
func GetRepo() Repo {
	return defRepo
}

type repoImpl struct{}

func (r repoImpl) MultiGetTasksStatus(ctx context.Context, uid string, buttons []*pb.Button) (map[int32]task.Task, error) {
	buttonIDs := lo.Map(buttons, func(btn *pb.Button, _ int) int32 {
		return btn.ButtonId
	})
	// 获取用户任务数据
	tds, err := user_task_data_repo.GetRepo().MultiGet(ctx, uid, buttonIDs)
	if err != nil {
		logger.Error(ctx, "MultiGetTasksStatus call user_task_data_repo.GetRepo().MultiGet err", zap.String("uid", uid), zap.Int32s("buttonIDs", buttonIDs), zap.Error(err))
		return nil, err
	}

	// 处理任务
	taskMM := lo.SliceToMap(buttons, func(btn *pb.Button) (int32, task.Task) {
		td := tds[btn.ButtonId] // 这里数据必然存在
		return btn.Task.TaskId, task.NewTask(uid, btn, td)
	})
	err = r.processTasks(ctx, uid, taskMM)
	if err != nil {
		logger.Error(ctx, "MultiGetTasksStatus call this.processTasks err", zap.Error(err))
		return nil, err
	}

	return taskMM, nil
}

func (r repoImpl) processTasks(ctx context.Context, uid string, taskMM map[int32]task.Task) error {
	// 更新周期
	for _, t := range taskMM {
		err := t.UpdatePeriod(ctx)
		if err != nil {
			logger.Error(ctx, "processTasks call t.UpdatePeriod err", zap.String("uid", uid), zap.Error(err))
			return err
		}
	}

	// 对不同任务分类
	taskTypeMM := make(map[pb.TaskType][]task.Task)
	for _, t := range taskMM {
		if !t.IsNeedQueryTaskProgress(ctx) {
			continue
		}
		btn := t.GetButton()
		taskTypeMM[btn.Task.TaskType] = append(taskTypeMM[btn.Task.TaskType], t)
	}

	// 批量刷新进度
	for tt, tasks := range taskTypeMM {
		if len(tasks) == 0 {
			continue
		}

		// 查询进度
		progress, err := task_progress.MultiQueryTaskProgress(ctx, tt, tasks)
		if err != nil {
			logger.Error(ctx, "processTasks call task_progress_repo.GetRepo().MultiQueryTaskProgress err", zap.String("uid", uid), zap.Int32("taskType", int32(tt)), zap.Error(err))
			return err
		}
		// 更新进度
		for i := range tasks {
			btnID := tasks[i].GetButton().ButtonId
			tasks[i].SetNewProgress(ctx, progress[btnID])
		}
	}

	// 持久化
	err := r.persistenceMultiTask(ctx, uid, taskMM)
	if err != nil {
		logger.Error(ctx, "processTasks call this.persistenceMultiTask err", zap.String("uid", uid), zap.Error(err))
		return err
	}
	return nil
}

func (r repoImpl) ClickOneButton(ctx context.Context, uid string, btn *pb.Button) (task.Task, error) {
	taskMM, err := r.MultiGetTasksStatus(ctx, uid, []*pb.Button{btn})
	if err != nil {
		logger.Error(ctx, "ClickOneButton call this.MultiGetTasksStatus err", zap.String("uid", uid), zap.Int32("buttonID", btn.ButtonId), zap.Error(err))
		return nil, err
	}
	t, ok := taskMM[btn.Task.TaskId]
	if !ok {
		// 理论上不会走到这个case
		logger.Error(ctx, "ClickOneButton abnormal. not found task.", zap.Int32("taskID", btn.Task.TaskId), zap.Any("btn", btn))
		return nil, fmt.Errorf("ClickOneButton abnormal. not found task. btnID=%d, taskID=%d", btn.ButtonId, btn.Task.TaskId)
	}

	err = t.ClickButton(ctx)
	if err != nil {
		logger.Error(ctx, "ClickOneButton call task.ClickButton err", zap.Int32("buttonID", btn.ButtonId), zap.Error(err))
		return nil, err
	}

	// click后需要重新持久化
	err = r.persistenceMultiTask(ctx, uid, map[int32]task.Task{t.GetButton().Task.TaskId: t})
	if err != nil {
		logger.Error(ctx, "ClickOneButton call this.persistenceOneTask err", zap.String("uid", uid), zap.Error(err))
		return t, err
	}
	return t, nil
}

// 批量领取任务奖励
func (r repoImpl) MultiFinishTasks(ctx context.Context, uid string, buttons []*pb.Button) (map[int32]task.Task, error) {
	taskMM, err := r.MultiGetTasksStatus(ctx, uid, buttons)
	if err != nil {
		logger.Error(ctx, "MultiFinishTasks call this.MultiGetTasksStatus err", zap.String("uid", uid), zap.Error(err))
		return nil, err
	}

	// 获取可领取奖励的任务
	retTasks := make(map[int32]task.Task)
	for _, t := range taskMM {
		if len(t.GetButton().Task.Prizes) == 0 {
			continue
		}
		hide := t.IsHide(ctx)
		if hide {
			continue
		}
		if t.GetUserTaskData().FinishStatus != pb.TaskFinishStatus_TASK_FINISH_STATUS_FINISHED {
			continue
		}

		// 只有可领取状态的任务参与计算
		retTasks[t.GetButton().Task.TaskId] = t
	}

	// 对已完成的任务进行点击
	fns := make([]func() error, len(retTasks))
	for _, t := range retTasks {
		t := t
		fns = append(fns, func() error {
			err = t.ClickButton(ctx)
			if err != nil {
				logger.Error(ctx, "MultiFinishTasks call task.ClickButton err", zap.String("uid", uid), zap.Int32("buttonID", t.GetButton().ButtonId), zap.Error(err))
				return err
			}
			return nil
		})
	}
	err = zapp.App().GetComponent().GetGPool().GoAndWait(fns...)
	if err != nil {
		return nil, err
	}
	if err != nil {
		logger.Error(ctx, "MultiFinishTasks multi call task.ClickButton err", zap.String("uid", uid), zap.Error(err))
		return nil, err
	}

	// 发货后需要重新持久化
	err = r.persistenceMultiTask(ctx, uid, retTasks)
	if err != nil {
		logger.Error(ctx, "MultiGetTasksStatus call this.persistenceMultiTask err", zap.Error(err))
		return nil, err
	}

	return retTasks, nil
}

// 持久化多个任务
func (repoImpl) persistenceMultiTask(ctx context.Context, uid string, taskMM map[int32]task.Task) error {
	// 批量持久化需要更新的任务数据
	needPersistenceUserTaskData := make(map[int32]*model.UserTaskData)
	for _, t := range taskMM {
		if t.IsNeedPersistence(ctx) {
			needPersistenceUserTaskData[t.GetButton().ButtonId] = t.GetUserTaskData()
		}
	}
	if len(needPersistenceUserTaskData) == 0 {
		return nil
	}
	err := user_task_data_repo.GetRepo().MultiUpdate(ctx, uid, needPersistenceUserTaskData)
	if err != nil {
		logger.Error(ctx, "persistenceMultiTask call user_task_data_repo.GetRepo().MultiUpdate err", zap.String("uid", uid), zap.Error(err))
		return err
	}
	return nil
}
