package util4iris

import (
	"github.com/kataras/iris"
	"github.com/lyoshur/util4iris/api"
	"github.com/lyoshur/util4iris/controller"
	"github.com/lyoshur/util4iris/form"
)

// api 返回状态码以及数据
//noinspection GoUnusedConst
const Success = api.Success

//noinspection GoUnusedConst
const Fail = api.Fail

//noinspection GoUnusedConst
const DataAnalysisFailMessage = api.DataAnalysisFailMessage

//noinspection GoUnusedConst
const ExecuteFailMessage = api.ExecuteFailMessage

//noinspection GoUnusedConst
const SuccessMessage = api.SuccessMessage

//noinspection GoUnusedConst
const NullData = api.NullData

type SimpleApi = api.SimpleApi

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
