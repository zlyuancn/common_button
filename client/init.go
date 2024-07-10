package client

import (
	"github.com/zly-app/cache/v2"
	"github.com/zly-app/component/redis"
	"github.com/zly-app/component/sqlx"
	"github.com/zly-app/zapp/core"

	"github.com/zlyuancn/common_button/conf"
)

var (
	sqlxCreator  sqlx.ISqlx
	redisCreator redis.IRedisCreator
	cacheCreator cache.ICacheCreator
)

func Init(app core.IApp) {
	sqlxCreator = sqlx.NewSqlx(app)
	redisCreator = redis.NewRedisCreator(app)
	cacheCreator = cache.NewCacheCreator(app)
}
func Close() {
	sqlxCreator.Close()
	redisCreator.Close()
	cacheCreator.Close()
}

func GetButtonSqlx() sqlx.Client {
	return sqlxCreator.GetSqlx(conf.Conf.ButtonSqlxName)
}
func GetUserTaskDataRedis() redis.UniversalClient {
	return redisCreator.GetRedis(conf.Conf.UserTaskDataRedisName)
}
func GetUserTaskDataCache() cache.ICache {
	return cacheCreator.GetCache(conf.Conf.UserTaskDataCacheName)
}
