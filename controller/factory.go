package controller

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/mvc"
	"sync"
)

// 路由的处理程序
type handler struct {
	// 处理链
	handlers context.Handlers
	// 控制器
	controller interface{}
}

// controller的工厂
var f *Factory

// 工厂结构体
// 此结构体用来维持controller层引用
type Factory struct {
	m map[string]handler
}

// 使用sync.Once保证factory单例
var once sync.Once

// 获取controller工厂
func GetFactory() *Factory {
	once.Do(func() {
		f = &Factory{}
		f.m = make(map[string]handler)
	})
	return f
}

// 注册controller
// 内部使用map进行维持数据结构
// 所以当注册的键重复时可能导致controller被替换，而发生错误
func (f *Factory) Register(s string, i interface{}, handlers ...context.Handler) {
	h := handler{}
	h.handlers = handlers
	h.controller = i
	f.m[s] = h
}

// 使用注册在内部的controller构建子路由
func (f *Factory) Build(application *iris.Application) {
	for k, v := range f.m {
		handlers := v.handlers
		controller := v.controller
		// 注册子路由
		mvc.New(application.Party(k, handlers...)).Handle(controller)
	}
}
