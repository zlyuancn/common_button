package hide_rule

import (
	"context"
	"fmt"

	"github.com/zly-app/zapp/logger"
	"go.uber.org/zap"

	"github.com/zlyuancn/common_button/model"
	"github.com/zlyuancn/common_button/pb"
)

type Rule interface {
	IsHide(ctx context.Context, btn *pb.Button, td *model.UserTaskData) (bool, error)
}

var rules = map[pb.TaskHideRule]Rule{
	pb.TaskHideRule_TASK_HIDE_RULE_NONE:     noneRule{},
	pb.TaskHideRule_TASK_HIDE_RULE_FINISHED: finishedRule{},
	pb.TaskHideRule_TASK_HIDE_RULE_RECEIVED: receivedRule{},
}

// 注册隐藏规则
func RegistryHideRule(t pb.TaskHideRule, rule Rule) {
	rules[t] = rule
}

// 检查是否隐藏
func CheckIsHide(ctx context.Context, btn *pb.Button, td *model.UserTaskData) (bool, error) {
	for _, t := range btn.Task.TaskHideRule {
		h, ok := rules[t]
		if !ok {
			err := fmt.Errorf("HideRule Type is invalid. t=%d", int(t))
			logger.Error(ctx, "CheckIsHide err", zap.Any("btn", btn), zap.Any("td", td), zap.Error(err))
			return false, err
		}

		hide, err := h.IsHide(ctx, btn, td)
		if err != nil {
			logger.Error(ctx, "CheckIsHide err", zap.Any("btn", btn), zap.Any("td", td), zap.Error(err))
			return false, err
		}
		if hide {
			return true, nil
		}
	}
	return false, nil
}

type noneRule struct{}

func (noneRule) IsHide(ctx context.Context, btn *pb.Button, td *model.UserTaskData) (bool, error) {
	return false, nil
}

type finishedRule struct{}

func (finishedRule) IsHide(ctx context.Context, btn *pb.Button, td *model.UserTaskData) (bool, error) {
	if td.FinishStatus != pb.TaskFinishStatus_TASK_FINISH_STATUS_UNFINISHED {
		return true, nil
	}
	return false, nil
}

type receivedRule struct{}

func (receivedRule) IsHide(ctx context.Context, btn *pb.Button, td *model.UserTaskData) (bool, error) {
	if td.FinishStatus == pb.TaskFinishStatus_TASK_FINISH_STATUS_RECEIVED {
		return true, nil
	}
	return false, nil
}
