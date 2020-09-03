package util4iris

import (
	"github.com/kataras/iris"
	"github.com/lyoshur/util4iris/controller"
	"github.com/lyoshur/util4iris/form"
)

// controller
type Factory = controller.Factory

//noinspection GoUnusedExportedFunction
func GetControllerFactory() *Factory {
	return controller.GetFactory()
}

// form
type SaveConfig = form.SaveConfig
type FileHolder = form.FileHolder
type File = form.File
type Helper = form.Helper

//noinspection GoUnusedExportedFunction
func GetFormHelper(ctx iris.Context) *Helper {
	return form.GetHelper(ctx)
}
