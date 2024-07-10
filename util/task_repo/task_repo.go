package task_repo

import (
	"context"
	"fmt"

	"github.com/mohae/deepcopy"
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
	// 批量渲染任务数据
	MultiRenderTasksStatus(ctx context.Context, buttons []*pb.Button) ([]*pb.Button, error)
}

var defRepo Repo = repoImpl{}

func SetRepo(repo Repo) {
	defRepo = repo
}
func GetRepo() Repo {
	return defRepo
}

type repoImpl struct{}

func (r repoImpl) MultiRenderTasksStatus(ctx context.Context, buttons []*pb.Button) ([]*pb.Button, error) {
	buttonIDs := make([]int32, 0)
	taskButtons := make([]*pb.Button, 0)
	for i, b := range buttons {
		if b.Task != nil {
			b = deepcopy.Copy(b).(*pb.Button) // 带任务的按钮会修改数据, 这里必须深拷贝
			buttons[i] = b                    // 回写

			buttonIDs = append(buttonIDs, b.ButtonId)
			taskButtons = append(taskButtons, b)
		}
	}
	if len(buttonIDs) == 0 {
		return buttons, nil
	}

	// 解析uid
	uid, err := user_repo.GetRepo().ParseUID(ctx)
	if err != nil {
		logger.Error(ctx, "MultiRenderTasksStatus call user_repo.GetRepo().ParseUID err", zap.Error(err))
		return nil, err
	}
	// 获取用户任务数据
	tds, err := user_task_data_repo.GetRepo().MultiGet(ctx, uid, buttonIDs)
	if err != nil {
		logger.Error(ctx, "MultiRenderTasksStatus call user_task_data_repo.GetRepo().MultiGet err", zap.String("uid", uid), zap.Int32s("buttonIDs", buttonIDs), zap.Error(err))
		return nil, err
	}

	// 处理任务
	taskMM := lo.SliceToMap(taskButtons, func(btn *pb.Button) (int32, task.Task) {
		td := tds[btn.Task.TaskId]
		return btn.Task.TaskId, task.NewTask(btn, td)
	})
	err = r.processTasks(ctx, uid, taskMM)
	if err != nil {
		logger.Error(ctx, "MultiRenderTasksStatus call this.processTasks err", zap.Error(err))
		return nil, err
	}

	// 过滤需要隐藏的按钮
	ret := make([]*pb.Button, 0, len(buttons))
	for _, btn := range buttons {
		if btn.Task == nil {
			ret = append(ret, btn)
			continue
		}

		t, ok := taskMM[btn.Task.TaskId]
		if !ok {
			// 理论上不会走到这个case
			logger.Error(ctx, "MultiRenderTasksStatus abnormal. not found task.", zap.Int32("taskID", btn.Task.TaskId), zap.Any("btn", btn))
			return nil, fmt.Errorf("MultiRenderTasksStatus abnormal. not found task. btnID=%d, taskID=%d", btn.ButtonId, btn.Task.TaskId)
		}

		hide, err := t.IsHide(ctx)
		if err != nil {
			logger.Error(ctx, "MultiRenderTasksStatus abnormal. check task IsHide err.", zap.Int32("taskID", btn.Task.TaskId), zap.Any("btn", btn), zap.Error(err))
			return nil, fmt.Errorf("MultiRenderTasksStatus abnormal. check task IsHide err. btnID=%d, taskID=%d, err=%v", btn.ButtonId, btn.Task.TaskId, err)
		}
		if !hide {
			ret = append(ret, btn)
		}
	}
	return ret, nil
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
	err := user_task_data_repo.GetRepo().MultiUpdate(ctx, uid, needPersistenceUserTaskData)
	if err != nil {
		logger.Error(ctx, "processTasks call user_task_data_repo.GetRepo().MultiUpdate err", zap.String("uid", uid), zap.Error(err))
		return err
	}
	return nil
}
