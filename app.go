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
	// Module 是模块。。
	// 作用是把繁琐的项目启动过程分离成独立的部分
	Module interface {
		Init(ac AppContext)
	}

	// ModuleFunc 方便把方法转换成Module
	ModuleFunc func(ac AppContext)

	// OnShutdown 在App停止前执行的方法
	OnShutdown func()

	// Job 是定时任务
	Job func(context.Context) error

	// AppContext 是模块初始化时可获取的资源和可调用的方法
	AppContext interface {
		// 注册GET路由
		GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc)
		// 注册POST路由
		POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc)
		// 注册中间件
		Use(middlewares ...echo.MiddlewareFunc)
		// 注册定时任务
		Schedule(expr string, job Job)
		// 发布事件
		Publish(topic string, payload string)
		// 订阅事件
		Subscribe(topic string, cb func(string))
		// 获取数据库连接
		GetDB() *gorm.DB
		// 获取redis连接
		GetRedis() *redis.Client
		// 日志方法
		Logf(format string, args ...interface{})
		// 注册资源，比如某些业务服务
		Provide(id string, obj interface{})
		// 获取资源
		Take(id string) interface{}
		// 注册关闭钩子，在http服务和定时任务服务停止后触发
		RegisterShutdown(hook OnShutdown)
	}
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

func NewApp(config Config) *App {
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

func (a *App) GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) {
	a.e.GET(path, h, m...)
}

func (a *App) POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) {
	a.e.POST(path, h, m...)
}

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

// 发布事件
func (a *App) Publish(topic string, payload string) {
	a.pubsub.Publish(topic, payload)
}

// 订阅事件
func (a *App) Subscribe(topic string, cb func(string)) {
	a.pubsub.Subscribe(topic, cb)
}

func (a *App) GetDB() *gorm.DB {
	return a.db
}

func (a *App) GetRedis() *redis.Client {
	return a.redisClient
}

func (a *App) RegisterModules(modules ...Module) {
	a.modules = modules
	for _, module := range modules {
		module.Init(a)
	}
}

func (a *App) Use(middlewares ...echo.MiddlewareFunc) {
	a.e.Use(middlewares...)
}

func (a *App) Provide(id string, obj interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.resource[id] = obj
}

func (a *App) Take(id string) interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.resource[id]
}

func (a *App) RegisterShutdown(hook OnShutdown) {
	a.shutdownHooks = append(a.shutdownHooks, hook)
}

// Start 启动HTTP和定时任务服务
// 通过返回的方法停止App
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
