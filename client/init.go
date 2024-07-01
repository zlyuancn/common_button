package client

import (
	"github.com/zly-app/component/redis"
	"github.com/zly-app/component/sqlx"
	"github.com/zly-app/zapp/core"

	"github.com/zlyuancn/common_button/conf"
)

var (
	sqlxCreator      sqlx.ISqlx
	ButtonSqlxClient sqlx.Client

	redisCreator        redis.IRedisCreator
	TaskDataRedisClient redis.UniversalClient
)

func Init(app core.IApp) {
	redisCreator = redis.NewRedisCreator(app)
	TaskDataRedisClient = redisCreator.GetRedis(conf.Conf.ButtonTaskDataRedisName)

	sqlxCreator = sqlx.NewSqlx(app)
	ButtonSqlxClient = sqlxCreator.GetSqlx(conf.Conf.ButtonSqlxName)
}
func Close() {
	sqlxCreator.Close()
	redisCreator.Close()
}
