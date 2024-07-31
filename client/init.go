package client

import (
	"github.com/zly-app/cache/v2"
	"github.com/zly-app/component/redis"
	"github.com/zly-app/component/sqlx"

	"github.com/zlyuancn/common_button/conf"
)

func GetButtonSqlx() sqlx.Client {
	return sqlx.GetClient(conf.Conf.ButtonSqlxName)
}
func GetUserTaskDataRedis() redis.UniversalClient {
	return redis.GetClient(conf.Conf.UserTaskDataRedisName)
}
func GetUserTaskDataCache() cache.ICache {
	return cache.GetCache(conf.Conf.UserTaskDataCacheName)
}
