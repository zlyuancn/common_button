package dao

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
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/logger"
	"github.com/zly-app/zapp/pkg/utils"
	"go.uber.org/zap"

	"github.com/zlyuancn/common_button/conf"
	"github.com/zlyuancn/common_button/dao/common_button"
	"github.com/zlyuancn/common_button/pb"
)

var ButtonTaskMap *loopload.LoopLoad[*buttonTaskMap]

type buttonTaskMap struct {
	ModuleSceneButtonMapping map[pb.ButtonModuleID]map[pb.ButtonSceneID][]*pb.Button // 业务模块的场景/页面映射按钮
	ButtonIDMapping          map[int32]*pb.Button                                    // 按钮id映射
	PrizeMapping             map[string]*pb.Prize                                    // 奖品映射
}

func startLoopLoad() {
	t := time.Duration(conf.Conf.ReloadButtonIntervalSec) * time.Second
	ButtonTaskMap = loopload.New("common_button", loadAllButtonTask, loopload.WithReloadTime(t))
}

func loadAllButtonTask(ctx context.Context) (*buttonTaskMap, error) {
	ctx = utils.Ctx.CloneContext(ctx) // 去掉ctx超时

	// 加载任务
	taskMM, prizeMM, err := loadTasksPB(ctx)
	if err != nil {
		logger.Error(ctx, "common_button call loadAllButtonTask call loadTasksPB err", zap.Error(err))
		return nil, err
	}

	// 加载按钮
	MSButtonMM, buttonMM, err := loadButtonPB(ctx, taskMM)
	if err != nil {
		logger.Error(ctx, "common_button call loadAllButtonTask call loadButtonPB err", zap.Error(err))
		return nil, err
	}

	// 组装
	ret := &buttonTaskMap{
		ModuleSceneButtonMapping: MSButtonMM,
		ButtonIDMapping:          buttonMM,
		PrizeMapping:             prizeMM,
	}
	return ret, nil
}

func loadTasksPB(ctx context.Context) (map[int32]*pb.Task, map[string]*pb.Prize, error) {
	// 加载任务
	taskList, err := common_button.LoadAllTask(ctx)
	if err != nil {
		logger.Error(ctx, "common_button call loadAllButtonTask call loadAllTask err", zap.Error(err))
		return nil, nil, err
	}
	// 解析奖品id
	prizeIDs := make([]string, 0)
	for _, t := range taskList {
		ids, err := parseTaskPrizeIDs(ctx, t)
		if err != nil {
			return nil, nil, err
		}
		prizeIDs = append(prizeIDs, ids...)
	}
	prizeIDs = lo.Uniq(prizeIDs)
	// 加载奖品数据
	prizeMM, err := loadPrizePB(ctx, prizeIDs)
	if err != nil {
		logger.Error(ctx, "common_button call loadPrizePB err", zap.Strings("prizeIDs", prizeIDs), zap.Error(err))
		return nil, nil, err
	}

	// 加载任务模板
	taskTemplateList, err := common_button.LoadAllTaskTemplate(ctx)
	if err != nil {
		logger.Error(ctx, "common_button call loadAllButtonTask call loadAllTaskTemplate err", zap.Error(err))
		return nil, nil, err
	}
	taskTemplateMM := lo.SliceToMap(taskTemplateList, func(t *common_button.TaskTemplateModel) (int32, *common_button.TaskTemplateModel) {
		return int32(t.ID), t
	})

	ret := make(map[int32]*pb.Task, len(taskList))
	for _, t := range taskList {
		tt, ok := taskTemplateMM[int32(t.TemplateID)]
		if !ok {
			logger.Error(ctx, "common_button 发现task使用了未定义的模板id", zap.Any("task", t))
			return nil, nil, fmt.Errorf("使用了未定义的模板id. taskID=%d templateID=%d", t.ID, t.TemplateID)
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

		// 解析隐藏规则
		hideRuleIsValid := true
		if t.HideRule != "" {
			ss := strings.Split(t.HideRule, ",")
			one.TaskHideRule = lo.FilterMap(ss, func(s string, _ int) (pb.TaskHideRule, bool) {
				ret, err := cast.ToInt32E(s)
				if err != nil {
					hideRuleIsValid = false
					logger.Error(ctx, "common_button 发现task的隐藏规则无效", zap.Any("task", t), zap.String("HideRule", s))
					return 0, false
				}
				return pb.TaskHideRule(ret), true
			})
			if !hideRuleIsValid {
				return nil, nil, fmt.Errorf("task的隐藏规则无效. taskID=%d HideRule=%s", t.ID, t.HideRule)
			}
		}

		// 解析奖品
		prizeIDs, err := parseTaskPrizeIDs(ctx, t)
		if err != nil {
			return nil, nil, err
		}
		prizes := make([]*pb.Prize, len(prizeIDs))
		for i, id := range prizeIDs {
			p, ok := prizeMM[id]
			if !ok {
				logger.Error(ctx, "common_button 发现task的某个奖品数据不存在", zap.Any("task", t), zap.String("prizeID", id))
				return nil, nil, fmt.Errorf("task的某个奖品数据不存在. taskID=%d, prizeID=%s", t.ID, id)
			}
			prizes[i] = p
		}
		one.Prizes = prizes

		ret[int32(t.ID)] = one
	}
	return ret, prizeMM, nil
}
func parseTaskPrizeIDs(ctx context.Context, t *common_button.TaskModel) ([]string, error) {
	if t.PrizeIds == "" {
		return nil, nil
	}

	ss := strings.Split(t.PrizeIds, ",")
	for i := 0; i < len(ss); i++ {
		if ss[i] == "" {
			logger.Error(ctx, "common_button 发现task的奖品id无效", zap.Any("task", t), zap.String("prizeIDs", t.PrizeIds))
			return nil, fmt.Errorf("common_button 发现task的奖品id无效. taskID=%d, prizeIDs=%s", t.ID, t.PrizeIds)
		}
	}
	return ss, nil
}
func loadPrizePB(ctx context.Context, prizeIDs []string) (map[string]*pb.Prize, error) {
	if len(prizeIDs) == 0 {
		return nil, nil
	}

	ch := make(chan *pb.Prize, len(prizeIDs))
	fns := make([]func() error, len(prizeIDs))
	for i := range prizeIDs {
		id := prizeIDs[i]
		fns = append(fns, func() error {
			v, err := prizeIDParse(ctx, id)
			if err != nil {
				return fmt.Errorf("common_button prizeIDParse err. prizeID=%s, err=%v", id, err)
			}
			ch <- v
			return nil
		})
	}
	err := zapp.App().GetComponent().GetGPool().GoAndWait(fns...)
	if err != nil {
		return nil, err
	}
	close(ch)

	ret := make(map[string]*pb.Prize, len(prizeIDs))
	for prize := range ch {
		ret[prize.PrizeId] = prize
	}
	return ret, nil
}

func genButtonPB(ctx context.Context, b *common_button.ButtonModel, taskMM map[int32]*pb.Task) (*pb.Button, error) {
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
func loadButtonPB(ctx context.Context, taskMM map[int32]*pb.Task) (
	map[pb.ButtonModuleID]map[pb.ButtonSceneID][]*pb.Button, map[int32]*pb.Button, error,
) {
	// 加载module
	moduleList, err := common_button.LoadAllModule(ctx)
	if err != nil {
		logger.Error(ctx, "common_button call loadAllButtonTask call loadAllModule err", zap.Error(err))
		return nil, nil, err
	}
	// 加载scene
	sceneList, err := common_button.LoadAllScene(ctx)
	if err != nil {
		logger.Error(ctx, "common_button call loadAllButtonTask call loadAllScene err", zap.Error(err))
		return nil, nil, err
	}
	// 加载按钮
	buttonList, err := common_button.LoadAllButton(ctx)
	if err != nil {
		logger.Error(ctx, "common_button call loadAllButtonTask call loadAllButton err", zap.Error(err))
		return nil, nil, err
	}
	sort.Slice(buttonList, func(i, j int) bool {
		if buttonList[i].SortValue != buttonList[j].SortValue {
			return buttonList[i].SortValue < buttonList[j].SortValue
		}
		return buttonList[i].Ctime.Unix() < buttonList[j].Ctime.Unix()
	})

	msButtonMM := make(map[pb.ButtonModuleID]map[pb.ButtonSceneID][]*pb.Button, len(moduleList))
	buttonMM := make(map[int32]*pb.Button, len(buttonList))
	for _, module := range moduleList {
		msButtonMM[pb.ButtonModuleID(module.ModuleID)] = make(map[pb.ButtonSceneID][]*pb.Button)
	}
	for _, scene := range sceneList {
		scenes, ok := msButtonMM[pb.ButtonModuleID(scene.ModuleID)]
		if !ok {
			logger.Error(ctx, "common_button 发现scene使用了未定义的module", zap.Any("scene", scene))
			return nil, nil, fmt.Errorf("scene使用了未定义的module. sceneID=%d, moduleID=%d", scene.SceneID, scene.ModuleID)
		}
		scenes[pb.ButtonSceneID(scene.SceneID)] = make([]*pb.Button, 0)
	}
	for _, b := range buttonList {
		scenes, ok := msButtonMM[pb.ButtonModuleID(b.ModuleID)]
		if !ok {
			logger.Error(ctx, "common_button 发现button使用了未定义的module", zap.Any("button", b))
			return nil, nil, fmt.Errorf("button使用了未定义的module. buttonID=%d, moduleID=%d", b.ID, b.ModuleID)
		}
		buttons, ok := scenes[pb.ButtonSceneID(b.SceneID)]
		if !ok {
			logger.Error(ctx, "common_button 发现button使用了未定义的scene", zap.Any("button", b))
			return nil, nil, fmt.Errorf("button使用了未定义的module. buttonID=%d, moduleID=%d, sceneID=%d", b.ID, b.ModuleID, b.SceneID)
		}

		bPB, err := genButtonPB(ctx, b, taskMM)
		if err != nil {
			logger.Error(ctx, "common_button button无法转为pb", zap.Any("button", b), zap.Error(err))
			return nil, nil, fmt.Errorf("button无法转为pb. buttonID=%d, err=%v", b.ID, err)
		}

		scenes[pb.ButtonSceneID(b.SceneID)] = append(buttons, bPB)
		buttonMM[int32(b.ID)] = bPB
	}
	return msButtonMM, buttonMM, nil
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
