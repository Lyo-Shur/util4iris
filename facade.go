package util4iris

import (
	"github.com/kataras/iris"
	"github.com/lyoshur/util4iris/api"
	"github.com/lyoshur/util4iris/controller"
	"github.com/lyoshur/util4iris/form"
)

// api 返回状态码以及数据
const Success = api.Success
const Fail = api.Fail
const DataAnalysisFailMessage = api.DataAnalysisFailMessage
const ExecuteFailMessage = api.ExecuteFailMessage
const SuccessMessage = api.SuccessMessage
const NullData = api.NullData

type SimpleApi = api.SimpleApi

// controller
type Factory = controller.Factory

func GetControllerFactory() *Factory {
	return controller.GetFactory()
}

// form
type SaveConfig = form.SaveConfig
type FileHolder = form.FileHolder
type File = form.File
type Helper = form.Helper

func GetFormHelper(ctx iris.Context) *Helper {
	return form.GetHelper(ctx)
}
