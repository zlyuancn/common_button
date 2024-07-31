package common_button

import (
	"context"

	"github.com/zly-app/grpc"
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/handler"
	"go.uber.org/zap"

	"github.com/zlyuancn/common_button/conf"
	"github.com/zlyuancn/common_button/dao/loopload"
	"github.com/zlyuancn/common_button/pb"
	"github.com/zlyuancn/common_button/view"
)

func init() {
	config.RegistryApolloNeedParseNamespace(conf.ConfigKey)

	loopload.StartLoopLoad()

	zapp.AddHandler(zapp.BeforeInitializeHandler, func(app core.IApp, handlerType handler.HandlerType) {
		err := app.GetConfig().Parse(conf.ConfigKey, &conf.Conf, true)
		if err != nil {
			app.Fatal("parse common_button config err", zap.Error(err))
		}
		conf.Conf.Check()
	})
	zapp.AddHandler(zapp.BeforeStartHandler, func(app core.IApp, handlerType handler.HandlerType) {
		pb.RegisterCommonButtonServiceServer(grpc.Server("common_button"), view.NewButtonService())

		grpcClient := pb.NewCommonButtonServiceClient(grpc.GetGatewayClientConn(conf.Conf.ButtonGrpcGatewayClientName))
		_ = pb.RegisterCommonButtonServiceHandlerClient(context.Background(), grpc.GetGatewayMux(), grpcClient)
	})
}
