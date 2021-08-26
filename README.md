# quick

## Installation

`go get github.com/hiwjd/quick`

## Example

```go
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/hiwjd/quick"
	"github.com/labstack/echo/v4"
)

func main() {
	app := quick.NewApp(quick.Config{
		APIAddr:  ":5555",
		MysqlDSN: "root:@/test?charset=utf8&parseTime=True&loc=Local",
		Redis: quick.Redis{
			Addr:     "127.0.0.1:6379",
			Password: "",
			DB:       0,
		},
	})
	app.RegisterModules(
		quick.ModuleFunc(demoModule),
	)
	close := app.Start()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	app.Logf("[INFO] Start closing")

	close()
}

type Comment struct {
	ID uint `gorm:"primaryKey"`
	Content string `gorm:"varchar(100)"`
	CreatedAt time.Time
}

func demoModule(ac quick.AppContext) {
	// register HTTP GET router
	ac.GET("/hello", func(c echo.Context) error {
		return c.String("world")
	})

	// register HTTP POST router
	ac.POST("/comment", func(c echo.Context) error {
		content := c.FormValue("content")
		comment := Comment{
			Content: content,
			CreatedAt: time.Now(),
		}
		if err := ac.GetDB().Save(&comment).Error; err != nil {
			return echo.NewHTTPError(500, "failed")
		}
		return comment
	})

	// register schedule job
	ac.Schedule("*/5 * * * * *", func(c context.Context) error {
		ac.Logf("job scheduled")

		// publish events
		ac.Publish("job-scheduled", "some payload")
	})

	// subscribe events
	ac.Subscribe("job-scheduled", func(payload string) {
		ac.Logf(payload) // some payload
	})

	// register shutdown hook
	ac.RegisterShutdown(func() {
		ac.Logf("shutdown...")
	})
}
```
