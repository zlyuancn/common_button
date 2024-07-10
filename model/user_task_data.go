package model

import (
	"github.com/zlyuancn/common_button/pb"
)

// 任务数据
type UserTaskData struct {
	PeriodMark   string              `json:"a,omitempty"` // 当前周期
	Progress     int32               `json:"b,omitempty"` // 任务进度
	FinishStatus pb.TaskFinishStatus `json:"c,omitempty"` // 任务完成状态
}
