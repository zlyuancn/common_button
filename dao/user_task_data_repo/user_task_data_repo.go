package user_task_data_repo

import (
	"context"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"github.com/zly-app/zapp/logger"
	"go.uber.org/zap"

	"github.com/zlyuancn/common_button/client"
	"github.com/zlyuancn/common_button/conf"
	"github.com/zlyuancn/common_button/model"
)

// 模板字符串
const (
	templateString_Uid      = "<uid>"
	templateString_ButtonID = "<btn_id>"
)

type Repo interface {
	// 批量获取任务数据
	MultiGet(ctx context.Context, uid string, buttonIDs []int32) (map[int32]*model.UserTaskData, error)
	// 批量更新任务数据
	MultiUpdate(ctx context.Context, uid string, data map[int32]*model.UserTaskData) error
}

var defRepo Repo = repoImpl{}

func SetRepo(repo Repo) {
	defRepo = repo
}
func GetRepo() Repo {
	return defRepo
}

type repoImpl struct{}

func (r repoImpl) MultiGet(ctx context.Context, uid string, buttonIDs []int32) (map[int32]*model.UserTaskData, error) {
	if len(buttonIDs) == 0 {
		return nil, nil
	}

	keys := lo.Map(buttonIDs, func(buttonID int32, _ int) string {
		return r.genTaskDataKey(uid, buttonID)
	})
	val, err := client.GetUserTaskDataRedis().MGet(ctx, keys...).Result()
	if err != nil {
		logger.Error(ctx, "MultiGet Call redis.MGet err", zap.String("uid", uid), zap.Strings("keys", keys), zap.Error(err))
		return nil, err
	}

	ret := make(map[int32]*model.UserTaskData, len(buttonIDs))
	for i, v := range val {
		id := buttonIDs[i]
		if v == nil {
			ret[id] = &model.UserTaskData{}
			continue
		}

		td := model.UserTaskData{}
		err := sonic.UnmarshalString(cast.ToString(v), &td)
		if err != nil {
			logger.Error(ctx, "MultiGet Call UnmarshalString err", zap.String("v", cast.ToString(v)), zap.Error(err))
			return nil, err
		}
		ret[id] = &td
	}
	return ret, nil
}

func (r repoImpl) MultiUpdate(ctx context.Context, uid string, data map[int32]*model.UserTaskData) error {
	if len(data) == 0 {
		return nil
	}

	values := make([]interface{}, len(data)*2)
	for id, td := range data {
		key := r.genTaskDataKey(uid, id)
		text, err := sonic.MarshalString(td)
		if err != nil {
			logger.Error(ctx, "MultiUpdate call MarshalString err", zap.String("uid", uid), zap.Int32("buttonID", id), zap.Any("td", td), zap.Error(err))
			return err
		}
		values = append(values, key, text)
	}
	err := client.GetUserTaskDataRedis().MSet(ctx, values...).Err()
	if err != nil {
		logger.Error(ctx, "MultiUpdate call MSet err", zap.String("uid", uid), zap.Any("values", values), zap.Error(err))
		return err
	}
	return nil
}

// 生成任务数据key
func (repoImpl) genTaskDataKey(uid string, buttonID int32) string {
	text := conf.Conf.UserTaskDataKeyFormat
	text = strings.ReplaceAll(text, templateString_Uid, uid)
	text = strings.ReplaceAll(text, templateString_ButtonID, cast.ToString(buttonID))
	return text
}
