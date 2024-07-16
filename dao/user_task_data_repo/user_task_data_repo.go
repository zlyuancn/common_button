package user_task_data_repo

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"github.com/zly-app/cache/v2"
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
	// 加用户锁
	LockUser(ctx context.Context, uid string) (unlock func(ctx context.Context), ok bool, err error)
	// 批量获取任务数据
	MultiGet(ctx context.Context, uid string, buttonIDs []int32) (map[int32]*model.UserTaskData, error)
	// 批量更新任务数据
	MultiUpdate(ctx context.Context, uid string, tds map[int32]*model.UserTaskData) error
}

var defRepo Repo = repoImpl{}

func SetRepo(repo Repo) {
	defRepo = repo
}
func GetRepo() Repo {
	return defRepo
}

type repoImpl struct{}

func (r repoImpl) LockUser(ctx context.Context, uid string) (func(ctx context.Context), bool, error) {
	startTime := time.Now().UnixNano()

	key := conf.Conf.UserOpLockKeyFormat
	key = strings.ReplaceAll(key, templateString_Uid, uid)
	expireTime := time.Duration(conf.Conf.UserOpLockTimeSec) * time.Second

	ok, err := client.GetUserTaskDataRedis().SetNX(ctx, key, 1, expireTime).Result()
	if err != nil {
		logger.Log.Error(ctx, "LockUser call redis.SetNx err", zap.String("uid", uid), zap.String("key", key), zap.Error(err))
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}

	var isUnlock int32 // unlock只能执行一次
	unlock := func(ctx context.Context) {
		if !atomic.CompareAndSwapInt32(&isUnlock, 0, 1) {
			return
		}
		nowTime := time.Now().UnixNano()
		if time.Duration(nowTime-startTime) > expireTime/2 { // 如果超过一半过期时间则不解锁
			return
		}
		err = client.GetUserTaskDataRedis().Del(ctx, key).Err()
		if err != nil {
			logger.Log.Error(ctx, "LockUser unlock err", zap.String("uid", uid), zap.String("key", key), zap.Error(err))
		}
	}
	return unlock, true, nil
}

func (r repoImpl) MultiGet(ctx context.Context, uid string, buttonIDs []int32) (map[int32]*model.UserTaskData, error) {
	if len(buttonIDs) == 0 {
		return nil, nil
	}
	ret := make(map[int32]*model.UserTaskData, len(buttonIDs))

	allKeys := lo.Map(buttonIDs, func(buttonID int32, _ int) string {
		return r.genTaskDataKey(uid, buttonID)
	})
	needQueryKeys := allKeys
	needQueryButtonIDs := buttonIDs

	// 使用缓存, 这里仅考虑使用了本地缓存, 从本地缓存获取数据无需并发
	if conf.Conf.UseUserTaskDataCache {
		needQueryKeys = make([]string, 0, len(buttonIDs))
		needQueryButtonIDs = make([]int32, 0, len(buttonIDs))
		td := model.UserTaskData{}
		for i := range allKeys {
			err := client.GetUserTaskDataCache().Get(ctx, allKeys[i], &td)
			if err == cache.ErrCacheMiss {
				needQueryKeys = append(needQueryKeys, allKeys[i])
				needQueryButtonIDs = append(needQueryButtonIDs, buttonIDs[i])
				continue
			}
			if err != nil {
				logger.Error(ctx, "MultiGet Call cache.Get err", zap.String("key", allKeys[i]), zap.Error(err))
				return nil, err
			}
			id := buttonIDs[i]
			ret[id] = &td
		}
		if len(needQueryKeys) == 0 {
			return ret, nil
		}
	}

	val, err := client.GetUserTaskDataRedis().MGet(ctx, needQueryKeys...).Result()
	if err != nil {
		logger.Error(ctx, "MultiGet Call redis.MGet err", zap.String("uid", uid), zap.Strings("allKeys", needQueryKeys), zap.Error(err))
		return nil, err
	}

	for i, v := range val {
		id := needQueryButtonIDs[i]
		td := model.UserTaskData{}
		if v != nil {
			err := sonic.UnmarshalString(cast.ToString(v), &td)
			if err != nil {
				logger.Error(ctx, "MultiGet Call UnmarshalString err", zap.String("v", cast.ToString(v)), zap.Error(err))
				return nil, err
			}
		}

		ret[id] = &td

		// 写入到缓存中
		if conf.Conf.UseUserTaskDataCache {
			err = client.GetUserTaskDataCache().Set(ctx, needQueryKeys[i], &td)
			if err != nil {
				logger.Error(ctx, "MultiGet Call cache.Set err", zap.String("key", needQueryKeys[i]), zap.Error(err))
				return nil, err
			}
		}
	}
	return ret, nil
}

func (r repoImpl) MultiUpdate(ctx context.Context, uid string, tds map[int32]*model.UserTaskData) error {
	if len(tds) == 0 {
		return nil
	}

	values := make([]interface{}, 0, len(tds)*2)
	for id, td := range tds {
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

	// 写入到缓存中, 这里仅考虑使用了本地缓存, 从本地缓存获取数据无需并发
	if conf.Conf.UseUserTaskDataCache {
		for id, td := range tds {
			key := r.genTaskDataKey(uid, id)
			err = client.GetUserTaskDataCache().Set(ctx, key, &td)
			if err != nil {
				logger.Error(ctx, "MultiUpdate Call cache.Set err", zap.String("key", key), zap.Error(err))
				return err
			}
		}
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
