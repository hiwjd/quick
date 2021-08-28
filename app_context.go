package quick

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type (
	// OnShutdown 在App停止前执行的方法
	OnShutdown func()

	// Job 是定时任务
	Job func(context.Context) error

	// AppContext 是模块初始化时可获取的资源和可调用的方法
	AppContext interface {
		// GET 注册HTTP GET路由
		GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc)
		// POST 注册HTTP POST路由
		POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc)
		// 注册中间件
		Use(middlewares ...echo.MiddlewareFunc)
		// Schedule 注册定时任务
		Schedule(expr string, job Job)
		// Publish 发布事件
		Publish(topic string, payload string)
		// Subscribe 订阅事件
		Subscribe(topic string, cb func(string))
		// GetDB 获取数据库连接实例
		GetDB() *gorm.DB
		// GetRedis 获取Redis连接实例
		GetRedis() *redis.Client
		// Logf 日志方法
		Logf(format string, args ...interface{})
		// Provide 提供资源，和Take配套使用
		Provide(id string, obj interface{})
		// Take 获取资源，即通过Provide提供的资源
		Take(id string) interface{}
		// RegisterShutdown 注册停止服务前调用的方法
		// 当服务停止时，会先停止HTTP服务、定时任务、事件系统，当这3者停止后，
		// 调用通过ReigsterShutdown注册的方法
		RegisterShutdown(hook OnShutdown)
	}
)
