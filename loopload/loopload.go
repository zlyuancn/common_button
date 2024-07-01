package loopload

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"github.com/zly-app/utils/loopload"
	"github.com/zly-app/zapp/logger"
	"go.uber.org/zap"

	"github.com/zlyuancn/common_button/conf"
	"github.com/zlyuancn/common_button/dao"
	"github.com/zlyuancn/common_button/pb"
)

var ButtonTaskMap *loopload.LoopLoad[*buttonTaskMap]

type buttonTaskMap struct {
	ModuleSceneButtonMapping map[pb.ButtonModuleID]map[pb.ButtonSceneID][]*pb.Button // 业务模块的场景/页面映射按钮
	ButtonIDMapping          map[int32]*pb.Button                                    // 按钮id映射
	PrizeMapping             map[int32]*pb.Prize                                     // 奖品映射
}

func Start() {
	t := time.Duration(conf.Conf.ReloadButtonIntervalSec) * time.Second
	ButtonTaskMap = loopload.New("common_button", loadAllButtonTask, loopload.WithReloadTime(t))
}

func loadAllButtonTask(ctx context.Context) (*buttonTaskMap, error) {
	// 加载module
	moduleList, err := dao.GetAllModule(ctx)
	if err != nil {
		logger.Error(ctx, "common_button call loadAllButtonTask call GetAllModule err: %v", err)
		return nil, err
	}
	// 加载scene
	sceneList, err := dao.GetAllScene(ctx)
	if err != nil {
		logger.Error(ctx, "common_button call loadAllButtonTask call GetAllScene err: %v", err)
		return nil, err
	}
	// 加载按钮
	buttonList, err := dao.GetAllButton(ctx)
	if err != nil {
		logger.Error(ctx, "common_button call loadAllButtonTask call GetAllButton err: %v", err)
		return nil, err
	}
	sort.Slice(buttonList, func(i, j int) bool {
		if buttonList[i].SortValue != buttonList[j].SortValue {
			return buttonList[i].SortValue < buttonList[j].SortValue
		}
		return buttonList[i].Ctime.Unix() < buttonList[j].Ctime.Unix()
	})

	// 加载任务模板
	taskTemplateList, err := dao.GetAllTaskTemplate(ctx)
	if err != nil {
		logger.Error(ctx, "common_button call loadAllButtonTask call GetAllTaskTemplate err: %v", err)
		return nil, err
	}
	taskTemplateMM := lo.SliceToMap(taskTemplateList, func(t *dao.TaskTemplateModel) (int32, *dao.TaskTemplateModel) {
		return int32(t.ID), t
	})
	// 加载任务
	taskList, err := dao.GetAllTask(ctx)
	if err != nil {
		logger.Error(ctx, "common_button call loadAllButtonTask call GetAllTask err: %v", err)
		return nil, err
	}
	taskMM, err := genTasksPB(ctx, taskList, taskTemplateMM)
	if err != nil {
		logger.Error(ctx, "common_button call loadAllButtonTask call genTasksPB err: %v", err)
		return nil, err
	}

	// 组装
	ret := &buttonTaskMap{
		ModuleSceneButtonMapping: make(map[pb.ButtonModuleID]map[pb.ButtonSceneID][]*pb.Button, len(moduleList)),
		ButtonIDMapping:          make(map[int32]*pb.Button, len(buttonList)),
		PrizeMapping:             make(map[int32]*pb.Prize),
	}

	for _, module := range moduleList {
		ret.ModuleSceneButtonMapping[pb.ButtonModuleID(module.ModuleID)] = make(map[pb.ButtonSceneID][]*pb.Button)
	}
	for _, scene := range sceneList {
		scenes, ok := ret.ModuleSceneButtonMapping[pb.ButtonModuleID(scene.ModuleID)]
		if !ok {
			logger.Error(ctx, "common_button 发现scene使用了未定义的module", zap.Any("scene", scene))
			return nil, fmt.Errorf("scene使用了未定义的module. sceneID=%d, moduleID=%d", scene.SceneID, scene.ModuleID)
		}
		scenes[pb.ButtonSceneID(scene.SceneID)] = make([]*pb.Button, 0)
	}
	for _, b := range buttonList {
		scenes, ok := ret.ModuleSceneButtonMapping[pb.ButtonModuleID(b.ModuleID)]
		if !ok {
			logger.Error(ctx, "common_button 发现button使用了未定义的module", zap.Any("button", b))
			return nil, fmt.Errorf("button使用了未定义的module. buttonID=%d, moduleID=%d", b.ID, b.ModuleID)
		}
		buttons, ok := scenes[pb.ButtonSceneID(b.SceneID)]
		if !ok {
			logger.Error(ctx, "common_button 发现button使用了未定义的scene", zap.Any("button", b))
			return nil, fmt.Errorf("button使用了未定义的module. buttonID=%d, moduleID=%d, sceneID=%d", b.ID, b.ModuleID, b.SceneID)
		}

		bPB, err := genButtonPB(ctx, b, taskMM)
		if err != nil {
			logger.Error(ctx, "common_button button无法转为pb", zap.Any("button", b), zap.Error(err))
			return nil, fmt.Errorf("button无法转为pb. buttonID=%d, err=%v", b.ID, err)
		}

		scenes[pb.ButtonSceneID(b.SceneID)] = append(buttons, bPB)
		ret.ButtonIDMapping[int32(b.ID)] = bPB
	}
	return ret, nil
}

func genButtonPB(ctx context.Context, b *dao.ButtonModel, taskMM map[int32]*pb.Task) (*pb.Button, error) {
	ret := &pb.Button{
		ModuleId:     pb.ButtonModuleID(b.ModuleID),
		SceneId:      pb.ButtonSceneID(b.SceneID),
		ButtonId:     int32(b.ID),
		ButtonTitle:  b.ButtonTitle,
		ButtonDesc:   b.ButtonDesc,
		Icon1:        b.Icon1,
		Icon2:        b.Icon2,
		Icon3:        b.Icon3,
		SortValue:    int32(b.SortValue),
		SkipValue:    b.SkipValue,
		SkipTitle:    b.SkipTitle,
		ButtonExtend: b.Extend,
		TaskState:    &pb.TaskState{},
	}

	if b.CommonTaskID != 0 {
		t, ok := taskMM[int32(b.CommonTaskID)]
		if !ok {
			logger.Error(ctx, "common_button 发现button使用了未定义的task", zap.Any("button", b))
			return nil, fmt.Errorf("button使用了未定义的task. buttonID=%d, taskID=%d", b.ID, b.CommonTaskID)
		}

		ret.Task = t
	}
	return ret, nil
}

func genTasksPB(ctx context.Context, taskList []*dao.TaskModel, taskTemplateMM map[int32]*dao.TaskTemplateModel) (map[int32]*pb.Task, error) {
	ret := make(map[int32]*pb.Task, len(taskList))
	for _, t := range taskList {
		tt, ok := taskTemplateMM[int32(t.TemplateID)]
		if !ok {
			logger.Error(ctx, "common_button 发现task使用了未定义的模板id", zap.Any("task", t))
			return nil, fmt.Errorf("使用了未定义的模板id. taskID=%d templateID=%d", t.ID, t.TemplateID)
		}

		one := &pb.Task{
			TaskId:             int32(t.ID),
			StartTime:          int32(t.StartTime.Unix()),
			EndTime:            int32(t.EndTime.Unix()),
			TaskTarget:         int32(t.TaskTarget),
			TaskExtend:         t.Extend,
			TaskPeriodType:     pb.TaskPeriodType(tt.PeriodType),
			TaskType:           pb.TaskType(tt.TaskType),
			TaskTemplateExtend: tt.Extend,
		}
		taskIsValid := true

		if t.HideRule != "" {
			ss := strings.Split(t.HideRule, ",")
			one.TaskHideRule = lo.FilterMap(ss, func(s string, _ int) (pb.TaskHideRule, bool) {
				ret, err := cast.ToInt32E(s)
				if err != nil {
					taskIsValid = false
					logger.Error(ctx, "common_button 发现task的隐藏规则无效", zap.Any("task", t), zap.String("HideRule", s))
					return 0, false
				}
				return pb.TaskHideRule(ret), true
			})
			if !taskIsValid {
				return nil, fmt.Errorf("task的隐藏规则无效. taskID=%d HideRule=%s", t.ID, t.HideRule)
			}
		}

		if t.PrizeIds != "" {
			ss := strings.Split(t.PrizeIds, ",")
			prizeIDs := lo.FilterMap(ss, func(s string, _ int) (int32, bool) {
				ret, err := cast.ToInt32E(s)
				if err != nil {
					taskIsValid = false
					logger.Error(ctx, "common_button 发现task的奖品id无效", zap.Any("task", t), zap.String("prizeID", s))
					return 0, false
				}
				return ret, true
			})
			if !taskIsValid {
				return nil, fmt.Errorf("task的奖品id无效. taskID=%d prizeIDs=%s", t.ID, t.PrizeIds)
			}

			prizes, err := genPrizePB(ctx, t, prizeIDs)
			if err != nil {
				logger.Error(ctx, "common_button task的奖品数据无法获取", zap.Any("task", t), zap.Error(err))
				return nil, fmt.Errorf("task的奖品数据无法获取. taskID=%d prizeIDs=%s err=%v", t.ID, t.PrizeIds, err)
			}
			one.Prizes = prizes
		}

		if taskIsValid {
			ret[int32(t.ID)] = one
		}
	}
	return ret, nil
}

func genPrizePB(ctx context.Context, t *dao.TaskModel, prizeIDs []int32) ([]*pb.Prize, error) {
	if len(prizeIDs) == 0 {
		return nil, nil
	}

	// todo 在此处注入奖品解析
	ret := make([]*pb.Prize, len(prizeIDs))
	for i := range prizeIDs {
		ret[i] = &pb.Prize{
			PrizeId:   prizeIDs[i],
			PrizeName: "",
			PrizeUrl:  "",
		}
	}
	return ret, nil
}

// 根据业务模块id和场景/页面id批量获取按钮, 场景/页面id为空则获取业务模块id下的所有按钮
func GetButtonsByModuleAndScene(ctx context.Context, moduleID pb.ButtonModuleID, sceneIDs []pb.ButtonSceneID) ([]*pb.Button, error) {
	btm := ButtonTaskMap.Get(ctx)
	scenes, ok := btm.ModuleSceneButtonMapping[moduleID]
	if !ok {
		return nil, fmt.Errorf("not found moduleID. moduleID=%d", moduleID)
	}

	ret := make([]*pb.Button, 0)

	// 获取所有
	if len(sceneIDs) == 0 {
		for _, bs := range scenes {
			ret = append(ret, bs...)
		}
	}

	for _, sid := range sceneIDs {
		bs, ok := scenes[sid]
		if !ok {
			return nil, fmt.Errorf("not found sceneID. sceneID=%d", sid)
		}
		ret = append(ret, bs...)
	}

	ret = deepcopy.Copy(ret).([]*pb.Button) // 此处必须深拷贝, 因为后面大概率会修改数据
	return ret, nil
}

// 根据按钮id获取按钮数据
func GetButtonByID(ctx context.Context, buttonID int32) (*pb.Button, error) {
	btm := ButtonTaskMap.Get(ctx)
	b, ok := btm.ButtonIDMapping[buttonID]
	if !ok {
		return nil, fmt.Errorf("not found buttonID. buttonID=%d", buttonID)
	}
	return b, nil
}
