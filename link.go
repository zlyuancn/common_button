package common_button

import (
	"github.com/zlyuancn/common_button/dao"
	"github.com/zlyuancn/common_button/dao/prize_repo"
	"github.com/zlyuancn/common_button/dao/user_task_data_repo"
	"github.com/zlyuancn/common_button/util/task_progress_repo"
	"github.com/zlyuancn/common_button/util/task_repo"
	"github.com/zlyuancn/common_button/util/user_repo"
)

// 设置按钮仓库
var SetButtonRepo = dao.SetButtonRepo

// 设置用户任务数据仓库
var SetUserTaskDataRepo = user_task_data_repo.SetRepo

// 设置任务仓库
var SetTaskRepo = task_repo.SetRepo

// 设置任务进度仓库
var SetTaskProgressRepo = task_progress_repo.SetRepo

// 设置奖品仓库
var SetPrizeRepo = prize_repo.SetRepo

// 设置用户仓库
var SetUserRepo = user_repo.SetRepo
