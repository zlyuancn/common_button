package common_button

import (
	"github.com/zly-app/grpc/gateway"
	"github.com/zly-app/grpc/server"
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
)

func WithService() zapp.Option {
	return zapp.WithCustomEnableService(func(app core.IApp, services []core.ServiceType) []core.ServiceType {
		services = addService(services, server.DefaultServiceType)
		services = addService(services, gateway.DefaultServiceType)
		return services
	})
}

func addService(services []core.ServiceType, t core.ServiceType) []core.ServiceType {
	for i := range services {
		if services[i] == t {
			return services
		}
	}
	services = append(services, t)
	return services
}
