package task_period

import (
	"context"
	"fmt"
	"time"

	"github.com/zlyuancn/zutils"

	"github.com/zlyuancn/common_button/model"
	"github.com/zlyuancn/common_button/pb"
)

type Period interface {
	// 生成周期标记
	GenPeriodMark(ctx context.Context, btn *pb.Button, td *model.UserTaskData) (string, error)
}

var periods = map[pb.TaskPeriodType]Period{
	pb.TaskPeriodType_TASK_PERIOD_TYPE_NONE:        nonePeriod{},
	pb.TaskPeriodType_TASK_PERIOD_TYPE_DAY_UTC8:    dayUtc8Period{},
	pb.TaskPeriodType_TASK_PERIOD_TYPE_WEEK_0_UTC8: week0Utc8Period{},
	pb.TaskPeriodType_TASK_PERIOD_TYPE_WEEK_1_UTC8: week1Utc8Period{},
	pb.TaskPeriodType_TASK_PERIOD_TYPE_MONTH_UTC8:  monthUtc8Period{},
}

// 注册任务周期
func RegistryPeriod(t pb.TaskPeriodType, period Period) {
	periods[t] = period
}

// 生成周期标记
func GenPeriodMark(ctx context.Context, btn *pb.Button, td *model.UserTaskData) (string, error) {
	p, ok := periods[btn.Task.TaskPeriodType]
	if !ok {
		return "", fmt.Errorf("TaskPeriod Type is invalid. t=%d", int(btn.Task.TaskPeriodType))
	}
	return p.GenPeriodMark(ctx, btn, td)
}

type nonePeriod struct{}

func (nonePeriod) GenPeriodMark(ctx context.Context, btn *pb.Button, td *model.UserTaskData) (string, error) {
	return "nonePeriod", nil
}

type dayUtc8Period struct{}

func (dayUtc8Period) GenPeriodMark(ctx context.Context, btn *pb.Button, td *model.UserTaskData) (string, error) {
	ret := zutils.Time(zutils.TZ.CSTTimeZone).TimeToTextOfLayout(time.Now(), zutils.T.LayoutDate)
	return ret, nil
}

type week0Utc8Period struct{}

func (week0Utc8Period) GenPeriodMark(ctx context.Context, btn *pb.Button, td *model.UserTaskData) (string, error) {
	t := zutils.Time(zutils.TZ.CSTTimeZone).GetWeekStartTimeOfWeek0(time.Now())
	ret := zutils.Time(zutils.TZ.CSTTimeZone).TimeToTextOfLayout(t, zutils.T.LayoutDate)
	return ret, nil
}

type week1Utc8Period struct{}

func (week1Utc8Period) GenPeriodMark(ctx context.Context, btn *pb.Button, td *model.UserTaskData) (string, error) {
	t := zutils.Time(zutils.TZ.CSTTimeZone).GetWeekStartTimeOfWeek1(time.Now())
	ret := zutils.Time(zutils.TZ.CSTTimeZone).TimeToTextOfLayout(t, zutils.T.LayoutDate)
	return ret, nil
}

type monthUtc8Period struct{}

func (monthUtc8Period) GenPeriodMark(ctx context.Context, btn *pb.Button, td *model.UserTaskData) (string, error) {
	ret := zutils.Time(zutils.TZ.CSTTimeZone).TimeToTextOfLayout(time.Now(), "2006-01")
	return ret, nil
}
