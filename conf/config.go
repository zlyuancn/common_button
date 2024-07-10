package conf

const ConfigKey = "common_button"

const (
	defButtonSqlxName              = "common_button"
	defReloadButtonIntervalSec     = 60
	defUserTaskDataRedisName       = "common_button"
	defUserTaskDataKeyFormat       = "{<uid>}:<btn_id>:common_button.user_task_data"
	defButtonGrpcGatewayClientName = "common_button"
	defUserTaskDataCacheName       = "common_button.user_task_data"
)

var Conf = Config{
	ButtonSqlxName:              defButtonSqlxName,
	ReloadButtonIntervalSec:     defReloadButtonIntervalSec,
	UserTaskDataRedisName:       defUserTaskDataRedisName,
	UserTaskDataKeyFormat:       defUserTaskDataKeyFormat,
	ButtonGrpcGatewayClientName: defButtonGrpcGatewayClientName,
	UserTaskDataCacheName:       defUserTaskDataCacheName,
}

type Config struct {
	ButtonSqlxName              string // 按钮的sqlx组件名
	ReloadButtonIntervalSec     int    // 重新加载按钮数据的间隔时间, 单位秒
	UserTaskDataRedisName       string // 用户任务数据的redis组件名
	UserTaskDataKeyFormat       string // 用户任务数据key格式化字符串
	ButtonGrpcGatewayClientName string // grpc网关客户端组件名
	UserTaskDataCacheName       string // 用户任务数据缓存组件名
}

func (conf *Config) Check() {
	if conf.ButtonSqlxName == "" {
		conf.ButtonSqlxName = defButtonSqlxName
	}
	if conf.ReloadButtonIntervalSec < 1 {
		conf.ReloadButtonIntervalSec = defReloadButtonIntervalSec
	}
	if conf.UserTaskDataRedisName == "" {
		conf.UserTaskDataRedisName = defUserTaskDataRedisName
	}
	if conf.UserTaskDataKeyFormat == "" {
		conf.UserTaskDataKeyFormat = defUserTaskDataKeyFormat
	}
	if conf.ButtonGrpcGatewayClientName == "" {
		conf.ButtonGrpcGatewayClientName = defButtonGrpcGatewayClientName
	}
	if conf.UserTaskDataCacheName == "" {
		conf.UserTaskDataCacheName = defUserTaskDataCacheName
	}
}
