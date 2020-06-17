package form

import (
	"github.com/Lyo-Shur/gutils"
	"github.com/kataras/iris"
)

// 表单参数帮助工具
type Helper struct {
	ctx iris.Context
}

// 获取表单参数帮助工具
func GetHelper(ctx iris.Context) *Helper {
	h := &Helper{}
	h.ctx = ctx
	return h
}

// 将表单中的值绑定到当前参数上
func (h *Helper) Binding(dest interface{}) error {
	params := make(map[string]string)
	for k, v := range h.ctx.FormValues() {
		params[k] = v[0]
	}
	return gutils.MapBindToStruct(params, dest)
}

// 取文件持有者
func (h *Helper) GetFileHolder() *FileHolder {
	fh := FileHolder{}
	// 上传的文件
	form := h.ctx.Request().MultipartForm
	if form != nil {
		fh.m = form.File
	}
	return &fh
}
