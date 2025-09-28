package common_button

import (
	"context"
	"time"

	"github.com/zlyuancn/common_button/client"
)

type TaskModel struct {
	ID         uint      `db:"id"`          // "任务id"
	ModuleID   uint      `db:"module_id"`   // "用于区分业务模块"
	SceneID    string    `db:"scene_id"`    // "业务下的场景/页面id"
	TemplateID uint      `db:"template_id"` // "模板id"
	StartTime  time.Time `db:"start_time"`  // "任务开始时间"
	EndTime    time.Time `db:"end_time"`    // "任务结束时间"
	TaskTarget uint      `db:"task_target"` // "任务目标"
	PrizeIds   string    `db:"prize_ids"`   // "奖品id列表，逗号分隔"
	HideRule   string    `db:"hide_rule"`   // "隐藏规则列表，逗号分隔：1=完成后隐藏 2=领奖后隐藏"
	Extend     string    `db:"extend"`      // "任务扩展, 一般用于存放任务模板无法确认的参数, 这些参数是运营决定的, 比如最近x天的x是多少"
	Remark     string    `db:"remark"`      // "备注"
}

// 按钮模型的字段
var taskSelectField = getModelSelectField(TaskModel{})

// 获取所有任务
func LoadAllTask(ctx context.Context) ([]*TaskModel, error) {
	var cond = `select ` + taskSelectField + ` from common_task where end_time > now()`

	var ret []*TaskModel
	err := client.GetButtonSqlx().Find(ctx, &ret, cond)
	return ret, err
}

type TaskTemplateModel struct {
	ID         uint   `db:"id"`          // "任务模板id"
	PeriodType int16  `db:"period_type"` // "任务周期：0=无周期 1=自然日 2=自然周"
	TaskType   int16  `db:"task_type"`   // "任务类型：1=跳转任务 2=签到"
	Extend     string `db:"extend"`      // "扩展数据, 一般用于存放任务模板数据的参数, 这些参数是开发者决定的, 比如第三方任务的id和secret"
	Remark     string `db:"remark"`      // "备注"
}

// 按钮模型的字段
var taskTemplateSelectField = getModelSelectField(TaskTemplateModel{})

// 获取所有任务模板
func LoadAllTaskTemplate(ctx context.Context) ([]*TaskTemplateModel, error) {
	var cond = `select ` + taskTemplateSelectField + ` from common_task_template`

	var ret []*TaskTemplateModel
	err := client.GetButtonSqlx().Find(ctx, &ret, cond)
	return ret, err
}
