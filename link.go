package common_button

import (
	"github.com/zlyuancn/common_button/dao"
	"github.com/zlyuancn/common_button/dao/prize_repo"
	"github.com/zlyuancn/common_button/dao/user_task_data_repo"
	"github.com/zlyuancn/common_button/util/hide_rule"
	"github.com/zlyuancn/common_button/util/task_period"
	"github.com/zlyuancn/common_button/util/task_progress"
	"github.com/zlyuancn/common_button/util/task_repo"
	"github.com/zlyuancn/common_button/util/user_repo"
)

// 设置按钮仓库
var SetButtonRepo = dao.SetButtonRepo

// 设置用户任务数据仓库
var SetUserTaskDataRepo = user_task_data_repo.SetRepo

// 设置任务仓库
var SetTaskRepo = task_repo.SetRepo

// 设置奖品仓库
var SetPrizeRepo = prize_repo.SetRepo

// 设置用户仓库
var SetUserRepo = user_repo.SetRepo

// 注册隐藏规则
var RegistryHideRule = hide_rule.RegistryHideRule

// 注册任务周期
var RegistryTaskPeriod = task_period.RegistryPeriod

// 注册任务进度解析函数
var RegistryTaskProgressHandle = task_progress.RegistryProgressHandle
