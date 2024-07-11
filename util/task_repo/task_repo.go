package task_repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/samber/lo"
	"github.com/zly-app/zapp/logger"
	"go.uber.org/zap"

	"github.com/zlyuancn/common_button/dao/user_task_data_repo"
	"github.com/zlyuancn/common_button/model"
	"github.com/zlyuancn/common_button/pb"
	"github.com/zlyuancn/common_button/util/task"
	"github.com/zlyuancn/common_button/util/task_progress"
	"github.com/zlyuancn/common_button/util/user_repo"
)

type Repo interface {
	// 批量获取任务状态
	MultiGetTasksStatus(ctx context.Context, buttons []*pb.Button) (map[int32]task.Task, error)
	// 获取单个任务状态
	GetOneTaskStatus(ctx context.Context, button *pb.Button) (task.Task, error)
	// 点击一个按钮扭转状态
	ClickOneButton(ctx context.Context, button *pb.Button) (task.Task, error)
}

var defRepo Repo = repoImpl{}

func SetRepo(repo Repo) {
	defRepo = repo
}
func GetRepo() Repo {
	return defRepo
}

type repoImpl struct{}

func (r repoImpl) MultiGetTasksStatus(ctx context.Context, buttons []*pb.Button) (map[int32]task.Task, error) {
	buttonIDs := make([]int32, 0)
	taskButtons := make([]*pb.Button, 0)
	for _, b := range buttons {
		if b.Task != nil {
			buttonIDs = append(buttonIDs, b.ButtonId)
			taskButtons = append(taskButtons, b)
		}
	}
	if len(buttonIDs) == 0 {
		return nil, nil
	}

	// 解析uid
	uid, err := user_repo.GetRepo().ParseUID(ctx)
	if err != nil {
		logger.Error(ctx, "MultiRenderTasksStatus call user_repo.GetRepo().ParseUID err", zap.Error(err))
		return nil, err
	}

	// 加锁
	unlock, ok, err := user_task_data_repo.GetRepo().LockUser(ctx, uid)
	if err != nil {
		logger.Error(ctx, "MultiRenderTasksStatus call user_task_data_repo.GetRepo().LockUser err", zap.Error(err))
		return nil, err
	}
	if !ok {
		err = errors.New("get lock err")
		logger.Error(ctx, "MultiRenderTasksStatus call user_task_data_repo.GetRepo().LockUser err", zap.Error(err))
		return nil, err
	}
	defer unlock(ctx)

	// 获取用户任务数据
	tds, err := user_task_data_repo.GetRepo().MultiGet(ctx, uid, buttonIDs)
	if err != nil {
		logger.Error(ctx, "MultiRenderTasksStatus call user_task_data_repo.GetRepo().MultiGet err", zap.String("uid", uid), zap.Int32s("buttonIDs", buttonIDs), zap.Error(err))
		return nil, err
	}

	// 处理任务
	taskMM := lo.SliceToMap(taskButtons, func(btn *pb.Button) (int32, task.Task) {
		td := tds[btn.ButtonId] // 这里数据必然存在
		return btn.Task.TaskId, task.NewTask(uid, btn, td)
	})
	err = r.processTasks(ctx, uid, taskMM)
	if err != nil {
		logger.Error(ctx, "MultiRenderTasksStatus call this.processTasks err", zap.Error(err))
		return nil, err
	}

	return taskMM, nil
}

func (repoImpl) processTasks(ctx context.Context, uid string, taskMM map[int32]task.Task) error {
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
			tasks[i].SetNewProgress(ctx, progress[i])
		}
	}

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
		logger.Error(ctx, "processTasks call user_task_data_repo.GetRepo().MultiUpdate err", zap.String("uid", uid), zap.Error(err))
		return err
	}
	return nil
}

func (r repoImpl) GetOneTaskStatus(ctx context.Context, btn *pb.Button) (task.Task, error) {
	if btn.Task == nil {
		return nil, fmt.Errorf("button not task. buttonID=%d", btn.ButtonId)
	}

	// 解析uid
	uid, err := user_repo.GetRepo().ParseUID(ctx)
	if err != nil {
		logger.Error(ctx, "GetOneTaskStatus call user_repo.GetRepo().ParseUID err", zap.Error(err))
		return nil, err
	}

	// 加锁
	unlock, ok, err := user_task_data_repo.GetRepo().LockUser(ctx, uid)
	if err != nil {
		logger.Error(ctx, "GetOneTaskStatus call user_task_data_repo.GetRepo().LockUser err", zap.Error(err))
		return nil, err
	}
	if !ok {
		err = errors.New("get lock err")
		logger.Error(ctx, "GetOneTaskStatus call user_task_data_repo.GetRepo().LockUser err", zap.Error(err))
		return nil, err
	}
	defer unlock(ctx)

	// 获取用户任务数据
	td, err := user_task_data_repo.GetRepo().Get(ctx, uid, btn.ButtonId)
	if err != nil {
		logger.Error(ctx, "GetOneTaskStatus call user_task_data_repo.GetRepo().Get err", zap.String("uid", uid), zap.Int32("buttonID", btn.ButtonId), zap.Error(err))
		return nil, err
	}

	t := task.NewTask(uid, btn, td)
	err = r.processOneTask(ctx, uid, t)
	if err != nil {
		logger.Error(ctx, "GetOneTaskStatus call this.processOneTask err", zap.Error(err))
		return nil, err
	}
	return t, nil
}

func (repoImpl) processOneTask(ctx context.Context, uid string, t task.Task) error {
	// 更新周期
	err := t.UpdatePeriod(ctx)
	if err != nil {
		logger.Error(ctx, "processOneTask call t.UpdatePeriod err", zap.String("uid", uid), zap.Error(err))
		return err
	}

	// 查询进度
	tt := t.GetButton().Task.TaskType
	progress, err := task_progress.MultiQueryTaskProgress(ctx, t.GetButton().Task.TaskType, []task.Task{t})
	if err != nil {
		logger.Error(ctx, "processOneTask call task_progress_repo.GetRepo().MultiQueryTaskProgress err", zap.String("uid", uid), zap.Int32("taskType", int32(tt)), zap.Error(err))
		return err
	}
	// 更新进度
	t.SetNewProgress(ctx, progress[0])

	// 持久化需要更新的任务数据
	if !t.IsNeedPersistence(ctx) {
		return nil
	}
	err = user_task_data_repo.GetRepo().Update(ctx, uid, t.GetButton().ButtonId, t.GetUserTaskData())
	if err != nil {
		logger.Error(ctx, "processOneTask call user_task_data_repo.GetRepo().Update err", zap.String("uid", uid), zap.Error(err))
		return err
	}
	return nil
}

func (r repoImpl) ClickOneButton(ctx context.Context, btn *pb.Button) (task.Task, error) {
	if btn.Task == nil {
		return nil, fmt.Errorf("button not task. buttonID=%d", btn.ButtonId)
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

	// 获取用户任务数据
	td, err := user_task_data_repo.GetRepo().Get(ctx, uid, btn.ButtonId)
	if err != nil {
		logger.Error(ctx, "ClickOneButton call user_task_data_repo.GetRepo().Get err", zap.String("uid", uid), zap.Int32("buttonID", btn.ButtonId), zap.Error(err))
		return nil, err
	}

	t := task.NewTask(uid, btn, td)
	err = r.processOneTask(ctx, uid, t)
	if err != nil {
		logger.Error(ctx, "ClickOneButton call this.processOneTask err", zap.Error(err))
		return nil, err
	}

	err = t.ClickButton(ctx)
	if err != nil {
		logger.Error(ctx, "ClickOneButton call task.ClickButton err", zap.Int32("buttonID", btn.ButtonId), zap.Error(err))
		return nil, err
	}

	// click后需要重新持久化
	if !t.IsNeedPersistence(ctx) {
		return t, nil
	}
	err = user_task_data_repo.GetRepo().Update(ctx, uid, t.GetButton().ButtonId, t.GetUserTaskData())
	if err != nil {
		logger.Error(ctx, "ClickOneButton call user_task_data_repo.GetRepo().Update err", zap.String("uid", uid), zap.Error(err))
		return t, err
	}
	return t, nil
}
