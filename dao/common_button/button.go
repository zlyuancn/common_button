package common_button

import (
	"context"
	"reflect"
	"strings"
	"time"

	"github.com/zlyuancn/common_button/client"
)

type ButtonModuleModel struct {
	ModuleID   uint   `db:"module_id"`   // 用于区分模块
	ModuleName string `db:"module_name"` // 模块名
}

// 获取所有业务模块
func LoadAllModule(ctx context.Context) ([]*ButtonModuleModel, error) {
	const cond = `select module_id,module_name from common_button_module`

	var ret []*ButtonModuleModel
	err := client.GetButtonSqlx().Find(ctx, &ret, cond)
	return ret, err
}

type ButtonSceneModel struct {
	ModuleID  uint   `db:"module_id"`  // 用于区分模块
	SceneID   string `db:"scene_id"`   // 子场景id
	SceneName string `db:"scene_name"` // 场景名
}

// 获取所有业务场景/页面
func LoadAllScene(ctx context.Context) ([]*ButtonSceneModel, error) {
	const cond = `select module_id,scene_id,scene_name from common_button_scene`

	var ret []*ButtonSceneModel
	err := client.GetButtonSqlx().Find(ctx, &ret, cond)
	return ret, err
}

type ButtonModel struct {
	ID           uint      `db:"id"`             // 按钮id
	ModuleID     uint      `db:"module_id"`      // 用于区分业务模块
	SceneID      string    `db:"scene_id"`       // 业务下的场景/页面id
	CommonTaskID uint      `db:"common_task_id"` // 通用任务id
	Enabled      byte      `db:"enabled"`        // 状态：0=未发布, 1=已发布
	SortValue    int       `db:"sort_value"`     // 顺序值. 正(数字小的在前), 排序值相同时以创建时间正序(新创建的在后)
	Extend       string    `db:"extend"`         // 扩展数据, json
	ButtonTitle  string    `db:"button_title"`   // 按钮标题
	ButtonDesc   string    `db:"button_desc"`    // 按钮描述/副标题
	Icon1        string    `db:"icon1"`          // 图片1
	Icon2        string    `db:"icon2"`          // 图片2
	Icon3        string    `db:"icon3"`          // 图片3
	SkipValue    string    `db:"skip_value"`     // 跳转地址
	SkipTitle    string    `db:"skip_title"`     // 跳转按钮标题
	Ctime        time.Time `db:"ctime"`          // 创建时间
}

// 按钮模型的字段
var buttonSelectField = getModelSelectField(ButtonModel{})

// 获取所有按钮
func LoadAllButton(ctx context.Context) ([]*ButtonModel, error) {
	var cond = `select ` + buttonSelectField + ` from common_button where enabled=1`

	var ret []*ButtonModel
	err := client.GetButtonSqlx().Find(ctx, &ret, cond)
	return ret, err
}

// 获取模型的select字段
func getModelSelectField(model interface{}) string {
	var selectAllFields []string
	rt := reflect.TypeOf(model)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	for _, field := range reflect.VisibleFields(rt) {
		// 拿到所有db字段
		if field.Tag.Get("db") != "" {
			selectAllFields = append(selectAllFields, field.Tag.Get("db"))
		}
	}
	return strings.Join(selectAllFields, ", ")
}
