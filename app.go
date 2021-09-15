package quick

import (
	"fmt"
	"io"
	"log"
	"os"

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
		Init(ac Context)
	}

	// ModuleFunc 方便把方法转换成Module
	ModuleFunc func(ac Context)

	// Migrator 迁移表
	Migrator func(db *gorm.DB) error
)

// Init 实现Module
func (mf ModuleFunc) Init(ac Context) {
	mf(ac)
}

type App struct {
	ac     *quickContext
	logger *log.Logger
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

	ac := &quickContext{}
	ac.config = config
	ac.logger = logger
	ac.c = c
	ac.db = initDB(config.MysqlDSN, logger)
	ac.redisClient = initRedis(config.Redis)
	ac.e = e
	ac.resource = make(map[string]interface{})
	ac.pubsub = newMemPubSub(ac.Logf)

	return &App{
		ac:     ac,
		logger: logger,
	}
}

// RegisterModules 注册模块，详情见Module
func (a *App) RegisterModules(modules ...Module) {
	a.ac.registerModules(modules...)
}

// Start 启动服务，并返回停止服务的方法
// 内部会根据配置启动HTTP服务、定时任务服务
func (a *App) Start() func() {
	return a.ac.start()
}

func (a *App) Logf(format string, args ...interface{}) {
	a.logger.Output(3, fmt.Sprintf(format, args...))
}

func (a *App) Context() Context {
	return a.ac
}

func (a *App) Migrate(migrators ...Migrator) {
	a.ac.migrate(migrators...)
}

// Provide 和AppContext.Provide拥有相同的功能，即注册资源到AppContext中
// 该方法返回Module，因此可以做为创建模块的快捷方式
// 比如这样使用: app.RegisterModules(quick.Provide("id-res1", obj))
func Provide(id string, obj interface{}) Module {
	return ModuleFunc(func(ac Context) {
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
