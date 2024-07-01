package common_button

import (
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/handler"
	"go.uber.org/zap"

	"github.com/zlyuancn/common_button/client"
	"github.com/zlyuancn/common_button/conf"
	"github.com/zlyuancn/common_button/loopload"
)

func init() {
	config.RegistryApolloNeedParseNamespace(conf.ConfigKey)

	zapp.AddHandler(zapp.BeforeInitializeHandler, func(app core.IApp, handlerType handler.HandlerType) {
		err := app.GetConfig().Parse(conf.ConfigKey, &conf.Conf, true)
		if err != nil {
			app.Fatal("parse common_button config err", zap.Error(err))
		}
		conf.Conf.Check()
	})
	zapp.AddHandler(zapp.AfterInitializeHandler, func(app core.IApp, handlerType handler.HandlerType) {
		client.Init(app)
		loopload.Start()
	})
	zapp.AddHandler(zapp.AfterExitHandler, func(app core.IApp, handlerType handler.HandlerType) {
		client.Close()
	})
}
