package quick

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/robfig/cron/v3"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type (
	// Module 是模块
	// 设定这个接口的作用是把繁杂的系统组成分离成独立的部分
	// 对具体什么是模块并没有任何约束，可以：
	//   提供HTTP服务
	//   执行定时任务
	//   独立存在的一段for循环
	//   仅仅打印一段话
	//   等等
	// 只要最终包装成Module接口传入RegisterModules方法即可
	// 可以查看samples中的项目找找感觉
	Module interface {
		Init(ac AppContext)
	}

	// ModuleFunc 方便把方法转换成Module
	ModuleFunc func(ac AppContext)
)

// Init 实现Module
func (mf ModuleFunc) Init(ac AppContext) {
	mf(ac)
}

type App struct {
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

// New 构造App
func New(config Config) *App {
	logWriter := initLoggerWriter(config.Log)
	logTimeFormt := "2006/01/02 15:04:05.00000"
	logger := log.New(logWriter, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format:           "${time_custom} [INFO] ${id} ${method} ${uri} ${latency_human} ${bytes_in} ${bytes_out} ${status} ${error}\n",
		CustomTimeFormat: logTimeFormt,
		Output:           logWriter,
	}))
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Skipper: func(c echo.Context) bool {
			return false
		},
		Generator: func() string {
			return uuid.NewString()
		},
		RequestIDHandler: func(c echo.Context, s string) {
			c.Set(echo.HeaderXRequestID, s)
		},
	}))
	e.HideBanner = true
	e.HTTPErrorHandler = NewCustomHTTPErrorHandler(e, func(format string, args ...interface{}) {
		logger.Printf(format, args...)
	})
	e.Validator = NewCustomValidator()

	cronParserOption := cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor
	c := cron.New(cron.WithLogger(cron.PrintfLogger(logger)), cron.WithParser(cron.NewParser(cronParserOption)))

	a := &App{}
	a.config = config
	a.logger = logger
	a.c = c
	a.db = initDB(config.MysqlDSN, logger)
	a.redisClient = initRedis(config.Redis)
	a.e = e
	a.resource = make(map[string]interface{})
	a.pubsub = newMemPubSub(a.Logf)

	return a
}

// GET 注册HTTP GET路由
func (a *App) GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) {
	a.e.GET(path, h, m...)
}

// POST 注册HTTP POST路由
func (a *App) POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) {
	a.e.POST(path, h, m...)
}

// Use 注册HTTP中间件
// 详细说明参考echo的文档 https://echo.labstack.com/middleware/#root-level-after-router
func (a *App) Use(middlewares ...echo.MiddlewareFunc) {
	a.e.Use(middlewares...)
}

// Schedule 注册定时任务
func (a *App) Schedule(expr string, job Job) {
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
func (a *App) Publish(topic string, payload string) {
	a.pubsub.Publish(topic, payload)
}

// Subscribe 订阅事件
func (a *App) Subscribe(topic string, cb func(string)) {
	a.pubsub.Subscribe(topic, cb)
}

// GetDB 获取数据库连接实例
func (a *App) GetDB() *gorm.DB {
	return a.db
}

// GetRedis 获取Redis连接实例
func (a *App) GetRedis() *redis.Client {
	return a.redisClient
}

// RegisterModules 注册模块，详情见Module
func (a *App) RegisterModules(modules ...Module) {
	for _, module := range modules {
		module.Init(a)
	}
	a.modules = append(a.modules, modules...)
}

// Provide 提供资源，和Take配套使用
func (a *App) Provide(id string, obj interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.resource[id] = obj
}

// Take 从AppContext中获取资源，即通过Provide提供的资源
func (a *App) Take(id string) interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.resource[id]
}

// RegisterShutdown 注册停止服务前调用的方法
// 当服务停止时，会先停止HTTP服务、定时任务、事件系统，当这3者停止后，
// 调用通过ReigsterShutdown注册的方法
func (a *App) RegisterShutdown(hook OnShutdown) {
	a.shutdownHooks = append(a.shutdownHooks, hook)
}

// Start 启动服务，并返回停止服务的方法
// 内部会根据配置启动HTTP服务、定时任务服务
func (a *App) Start() func() {
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

func (a *App) Logf(format string, args ...interface{}) {
	a.logger.Output(3, fmt.Sprintf(format, args...))
}

// Provide 和AppContext.Provide拥有相同的功能，即注册资源到AppContext中
// 该方法返回Module，因此可以做为创建模块的快捷方式
// 比如这样使用: app.RegisterModules(quick.Provide("id-res1", obj))
func Provide(id string, obj interface{}) Module {
	return ModuleFunc(func(ac AppContext) {
		ac.Provide(id, obj)
	})
}

func initDB(dsn string, writer logger.Writer) *gorm.DB {
	var db *gorm.DB
	var err error
	if dsn != "" {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
			},
			Logger: logger.New(writer, logger.Config{LogLevel: logger.Info}),
		})
		if err != nil {
			panic("Failed Open Database: " + err.Error())
		}
	}
	return db
}

func initRedis(cfg Redis) *redis.Client {
	var redisClient *redis.Client
	if cfg.Addr != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		})
	}
	return redisClient
}

func initLoggerWriter(cfg Log) io.Writer {
	var w io.Writer
	switch cfg.Output {
	case "stdout":
		w = os.Stdout
	default:
		w = &lumberjack.Logger{
			Filename:   cfg.Output,
			MaxSize:    cfg.MaxSize, // megabytes
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,   // days
			Compress:   cfg.Compress, // disabled by default
		}
	}
	return w
}
