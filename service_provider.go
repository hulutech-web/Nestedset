package nestedset

import (
	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "nestedset"

var App foundation.Application

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	App = app
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		//返回一个实例
		return &Nestedset{}, nil
	})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {

}
