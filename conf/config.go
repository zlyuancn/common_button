package conf

const ConfigKey = "common_button"

const (
	defButtonSqlxName              = "common_button"
	defReloadButtonIntervalSec     = 60
	defUserTaskDataRedisName       = "common_button"
	defUserTaskDataKeyFormat       = "common_button.user_task_data:{<uid>}:<btn_id>"
	defUseUserTaskDataCache        = true
	defUserTaskDataCacheName       = "common_button.user_task_data"
	defUserOpLockKeyFormat         = "common_button.user_op_lock:{<uid>}"
	defUserOpLockTimeSec           = 10
	defButtonGrpcGatewayClientName = "common_button"
)

var Conf = Config{
	ButtonSqlxName:              defButtonSqlxName,
	ReloadButtonIntervalSec:     defReloadButtonIntervalSec,
	UserTaskDataRedisName:       defUserTaskDataRedisName,
	UserTaskDataKeyFormat:       defUserTaskDataKeyFormat,
	UseUserTaskDataCache:        defUseUserTaskDataCache,
	UserTaskDataCacheName:       defUserTaskDataCacheName,
	UserOpLockKeyFormat:         defUserOpLockKeyFormat,
	UserOpLockTimeSec:           defUserOpLockTimeSec,
	ButtonGrpcGatewayClientName: defButtonGrpcGatewayClientName,
}

type Config struct {
	ButtonSqlxName              string // 按钮的sqlx组件名
	ReloadButtonIntervalSec     int    // 重新加载按钮数据的间隔时间, 单位秒
	UserTaskDataRedisName       string // 用户任务数据的redis组件名
	UserTaskDataKeyFormat       string // 用户任务数据key格式化字符串
	UseUserTaskDataCache        bool   // 是否使用用户数据缓存. 注意, 在使用分布式系统的情况下, 开启缓存注意将同一个用户的请求分配到同一个节点中
	UserTaskDataCacheName       string // 用户任务数据缓存组件名
	UserOpLockKeyFormat         string // 用户操作加锁key格式化字符串
	UserOpLockTimeSec           int64  // 用户操作加锁时间, 单位秒, 在redis中如果操作时间小于其一半时间会调用unlock解锁否则只能等待自动过期
	ButtonGrpcGatewayClientName string // grpc网关客户端组件名
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
	if conf.UserTaskDataCacheName == "" {
		conf.UserTaskDataCacheName = defUserTaskDataCacheName
	}
	if conf.UserOpLockKeyFormat == "" {
		conf.UserOpLockKeyFormat = defUserOpLockKeyFormat
	}
	if conf.UserOpLockTimeSec < 1 {
		conf.UserOpLockTimeSec = defUserOpLockTimeSec
	}
	if conf.ButtonGrpcGatewayClientName == "" {
		conf.ButtonGrpcGatewayClientName = defButtonGrpcGatewayClientName
	}
}
