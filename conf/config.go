package conf

const ConfigKey = "common_button"

const (
	defButtonSqlxName          = "common_button"
	defReloadButtonIntervalSec = 60

	defButtonTaskDataRedisName = "common_button"
)

var Conf = Config{
	ButtonSqlxName:          defButtonSqlxName,
	ReloadButtonIntervalSec: defReloadButtonIntervalSec,

	ButtonTaskDataRedisName: defButtonTaskDataRedisName,
}

type Config struct {
	ButtonSqlxName          string // 按钮的sqlx组件名
	ReloadButtonIntervalSec int    // 重新加载按钮数据的间隔时间

	ButtonTaskDataRedisName string // 按钮任务数据的redis组件名
}

func (conf *Config) Check() {
	if conf.ButtonSqlxName == "" {
		conf.ButtonSqlxName = defButtonSqlxName
	}
	if conf.ReloadButtonIntervalSec < 1 {
		conf.ReloadButtonIntervalSec = defReloadButtonIntervalSec
	}
}
