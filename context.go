package quick

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo/v4"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type (
	// OnShutdown 在App停止前执行的方法
	OnShutdown func()

	// Job 是定时任务
	Job func(context.Context) error

	// Context 是模块初始化时可获取的资源和可调用的方法
	Context interface {
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

type quickContext struct {
	muModule      sync.Mutex
	mu            sync.RWMutex
	config        Config
	logger        *log.Logger
	c             *cron.Cron
	e             *echo.Echo
	db            *gorm.DB
	redisClient   *redis.Client
	resource      map[string]interface{}
	modules       []Module
	shutdownHooks []OnShutdown
	pubsub        PubSub
}

// GET 注册HTTP GET路由
func (a *quickContext) GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) {
	a.e.GET(path, h, m...)
}

// POST 注册HTTP POST路由
func (a *quickContext) POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) {
	a.e.POST(path, h, m...)
}

// Use 注册HTTP中间件
// 详细说明参考echo的文档 https://echo.labstack.com/middleware/#root-level-after-router
func (a *quickContext) Use(middlewares ...echo.MiddlewareFunc) {
	a.e.Use(middlewares...)
}

// Schedule 注册定时任务
func (a *quickContext) Schedule(expr string, job Job) {
	fn := func() {
		ctx := context.Background()
		if err := job(ctx); err != nil {
			a.Logf("[ERROR] Cron Job Execute Failed: %s", err.Error())
		}
	}
	job0 := cron.NewChain(cron.DelayIfStillRunning(cron.PrintfLogger(a.logger))).Then((cron.FuncJob(fn)))

	entryID, err := a.c.AddJob(expr, job0)
	if err != nil {
		a.Logf("[ERROR] Cron Job Add Failed: %s, expr: %s", err.Error(), expr)
	} else {
		a.Logf("[INFO] Cron Job Add Success: %d", entryID)
	}
}

// Publish 发布事件
func (a *quickContext) Publish(topic string, payload string) {
	a.pubsub.Publish(topic, payload)
}

// Subscribe 订阅事件
func (a *quickContext) Subscribe(topic string, cb func(string)) {
	a.pubsub.Subscribe(topic, cb)
}

// GetDB 获取数据库连接实例
func (a *quickContext) GetDB() *gorm.DB {
	return a.db
}

// GetRedis 获取Redis连接实例
func (a *quickContext) GetRedis() *redis.Client {
	return a.redisClient
}

// Provide 提供资源，和Take配套使用
func (a *quickContext) Provide(id string, obj interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.resource[id] = obj
}

// Take 从quickContext中获取资源，即通过Provide提供的资源
func (a *quickContext) Take(id string) interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.resource[id]
}

// RegisterShutdown 注册停止服务前调用的方法
// 当服务停止时，会先停止HTTP服务、定时任务、事件系统，当这3者停止后，
// 调用通过ReigsterShutdown注册的方法
func (a *quickContext) RegisterShutdown(hook OnShutdown) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.shutdownHooks = append(a.shutdownHooks, hook)
}

func (a *quickContext) Logf(format string, args ...interface{}) {
	a.logger.Output(2, fmt.Sprintf(format, args...))
}

// registerModules 注册模块，详情见Module
func (a *quickContext) registerModules(modules ...Module) {
	a.muModule.Lock()
	defer a.muModule.Unlock()
	for _, module := range modules {
		module.Init(a)
	}
	a.modules = append(a.modules, modules...)
}

func (a *quickContext) migrate(migrators ...Migrator) {
	for _, migrator := range migrators {
		if err := migrator(a.db); err != nil {
			panic(fmt.Sprintf("Migrate failed: %s\n", err.Error()))
		}
	}
}

// start 启动服务，并返回停止服务的方法
// 内部会根据配置启动HTTP服务、定时任务服务
func (a *quickContext) start() func() {
	a.c.Start()
	go func() {
		if err := a.e.Start(a.config.APIAddr); err != nil && err != http.ErrServerClosed {
			a.Logf("[ERROR] Echo Start Failed: %s", err.Error())
			panic(err.Error())
		}
	}()

	return func() {
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			cCtx := a.c.Stop()
			<-cCtx.Done()
			a.Logf("[INFO] Cron Stopped")
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := a.e.Shutdown(ctx); err != nil {
				a.Logf("[ERROR] Echo Shutdown Failed: %s", err.Error())
			}
			a.Logf("[INFO] Echo Stopped")
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			a.pubsub.Close()
			a.Logf("[INFO] PubSub Stopped")
		}()

		wg.Wait()
		for _, hook := range a.shutdownHooks {
			func() {
				defer func() {
					if err := recover(); err != nil {
						a.Logf("Shutdown Hook Triggered Error: %#v", err)
					}
				}()
				hook()
			}()
		}
		a.Logf("[INFO] Stopped")
	}
}
